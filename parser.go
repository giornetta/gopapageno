package gopapageno

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"runtime/debug"
	"runtime/pprof"
)

var (
	discardLogger = log.New(io.Discard, "", 0)
)

type Rule struct {
	Lhs TokenType
	Rhs []TokenType
}

type ParserFunc func(rule uint16, lhs *Token, rhs []*Token, thread int)

// A ParseStrategy defines which kind of algorithm should be executed
// when collecting and running multiple parsing passes.
type ParseStrategy uint8

const (
	// StratSweep will run a single serial pass after combining data from the first `n` parallel runs.
	StratSweep ParseStrategy = iota
	// StratParallel will combine adjacent parsing results and recursively run `n-1` parallel runs until one stack remains.
	StratParallel
)

type Parser struct {
	Lexer *Lexer

	NumTerminals    uint16
	NumNonterminals uint16

	MaxRHSLength    int
	Rules           []Rule
	CompressedRules []uint16

	PrecedenceMatrix          [][]Precedence
	BitPackedPrecedenceMatrix []uint64

	Func ParserFunc

	PreallocFunc PreallocFunc

	concurrency int
	strategy    ParseStrategy

	logger *log.Logger

	cpuProfileWriter io.Writer
	memProfileWriter io.Writer
}

func (p *Parser) Concurrency() int {
	return p.concurrency
}

type ParserOpt func(p *Parser)

func WithConcurrency(n int) ParserOpt {
	return func(p *Parser) {
		if n <= 0 {
			n = 1
		}

		p.concurrency = n
	}
}

func WithLogging(logger *log.Logger) ParserOpt {
	return func(p *Parser) {
		if logger == nil {
			logger = discardLogger
		}

		p.logger = logger
	}
}

func WithCPUProfiling(w io.Writer) ParserOpt {
	return func(p *Parser) {
		p.cpuProfileWriter = w
	}
}

func WithMemoryProfiling(w io.Writer) ParserOpt {
	return func(p *Parser) {
		p.memProfileWriter = w
	}
}

func WithPreallocFunc(fn PreallocFunc) ParserOpt {
	return func(p *Parser) {
		p.PreallocFunc = fn
	}
}

func WithStrategy(strat ParseStrategy) ParserOpt {
	return func(p *Parser) {
		p.strategy = strat
	}
}

func NewParser(
	lexer *Lexer,
	numTerminals uint16, numNonterminals uint16, maxRHSLength int,
	rules []Rule, compressedRules []uint16,
	precedenceMatrix [][]Precedence, bitPackedPrecedenceMatrix []uint64,
	fn ParserFunc,
	opts ...ParserOpt,
) *Parser {
	parser := &Parser{
		Lexer:                     lexer,
		NumTerminals:              numTerminals,
		NumNonterminals:           numNonterminals,
		MaxRHSLength:              maxRHSLength,
		Rules:                     rules,
		CompressedRules:           compressedRules,
		PrecedenceMatrix:          precedenceMatrix,
		BitPackedPrecedenceMatrix: bitPackedPrecedenceMatrix,
		Func:                      fn,
		concurrency:               1,
		strategy:                  StratSweep,
		logger:                    discardLogger,
		cpuProfileWriter:          nil,
		memProfileWriter:          nil,
	}

	for _, opt := range opts {
		opt(parser)
	}

	return parser
}

func (p *Parser) Parse(ctx context.Context, src []byte) (*Token, error) {
	// Profiling
	if p.cpuProfileWriter != nil && p.cpuProfileWriter != io.Discard {
		if err := pprof.StartCPUProfile(p.cpuProfileWriter); err != nil {
			log.Printf("could not start CPU profiling: %v", err)
		}

		defer func() {
			if p.memProfileWriter != nil && p.memProfileWriter != io.Discard {
				if err := pprof.WriteHeapProfile(p.memProfileWriter); err != nil {
					log.Printf("Could not write memory profile: %v", err)
				}
			}

			pprof.StopCPUProfile()
		}()
	}

	// Run Prealloc Functions
	if p.Lexer.PreallocFunc != nil {
		p.Lexer.PreallocFunc(len(src), p.concurrency)
	}

	if p.PreallocFunc != nil {
		p.PreallocFunc(len(src), p.concurrency)
	}

	scanner := p.Lexer.Scanner(src, ScannerWithConcurrency(p.concurrency))

	srcLen := len(src)
	avgCharsPerToken := 12.5

	stackPoolBaseSize := math.Ceil(float64(srcLen) / avgCharsPerToken / float64(stackSize) / float64(p.concurrency))
	stackPtrPoolBaseSize := math.Ceil(float64(srcLen) / avgCharsPerToken / float64(pointerStackSize) / float64(p.concurrency))

	pools := make([]*Pool[stack[Token]], p.concurrency)
	ptrPools := make([]*Pool[tokenPointerStack], p.concurrency)

	for thread := 0; thread < p.concurrency; thread++ {
		pools[thread] = NewPool[stack[Token]](int(stackPoolBaseSize * 0.8))
		ptrPools[thread] = NewPool[tokenPointerStack](int(stackPtrPoolBaseSize))
	}

	stackPoolFinalPass := NewPool[stack[Token]](int(math.Ceil(stackPoolBaseSize * 0.1 * float64(p.concurrency))))
	stackPoolNewNonterminalsFinalPass := NewPool[stack[Token]](int(math.Ceil(stackPoolBaseSize * 0.05 * float64(p.concurrency))))
	stackPtrPoolFinalPass := NewPool[tokenPointerStack](int(math.Ceil(stackPtrPoolBaseSize * 0.1)))

	// TODO: Investigate this section better.
	// Old code forced a GC Run to occur, so that it would - hopefully - stop GCs from happening again during computation.
	// However a GC run can still be very slow.
	// runtime.GC()

	// This new version stops the GC from running entirely.
	debug.SetGCPercent(-1)

	// Deferring this will cause the GC to still run at the end of computation...
	// defer debug.SetGCPercent(1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tokens, err := scanner.Lex(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not lex: %w", err)
	}
	// If there are not enough stacks in the input, reduce the number of threads.
	// The input is split by splitting stacks, not stack contents.
	if tokens.NumStacks() < p.concurrency {
		p.concurrency = tokens.NumStacks()

		p.logger.Printf("Not enough stacks in lexer output, lowering parser concurrency to %d", p.concurrency)
	}

	tokensLists, err := tokens.Split(p.concurrency)
	if err != nil {
		return nil, err
	}

	resultCh := make(chan parseResult)
	errCh := make(chan error, 1)

	workers := make([]*parserWorker, p.concurrency)
	for thread := 0; thread < p.concurrency; thread++ {
		var nextToken *Token

		// If the thread is not the last, also take the first token of the next stack
		if thread < p.concurrency-1 {
			nextInputListIter := tokensLists[thread+1].HeadIterator()
			nextToken = nextInputListIter.Next()
		}

		workers[thread] = &parserWorker{
			parser: p,
			id:     thread,

			newNTList: NewListOfStacks[Token](pools[thread]),
			stack:     newListOfTokenPointerStacks(ptrPools[thread]),
		}

		go workers[thread].parse(ctx,
			tokensLists[thread],
			nextToken,
			false,
			resultCh,
			errCh)
	}

	parseResults := make([]*listOfTokenPointerStacks, p.concurrency)
	completed := 0

	for completed < p.concurrency {
		select {
		case result := <-resultCh:
			parseResults[result.threadNum] = result.stack
			completed++
		case err := <-errCh:
			cancel()
			return nil, err
		}
	}

	if p.strategy == StratSweep {
		//If the number of threads is greater than one, a final pass is required
		if p.concurrency > 1 {
			//Create the final input by joining together the stacks from the previous step
			finalPassInput := NewListOfStacks[Token](stackPoolFinalPass)

			for i := 0; i < p.concurrency; i++ {
				iterator := parseResults[i].HeadIterator()

				//Ignore the first token
				iterator.Next()

				for token := iterator.Next(); token != nil; token = iterator.Next() {
					finalPassInput.Push(*token)
				}
			}

			workers[0].newNTList = NewListOfStacks[Token](stackPoolNewNonterminalsFinalPass)
			workers[0].stack = newListOfTokenPointerStacks(stackPtrPoolFinalPass)

			p.concurrency = 1

			go workers[0].parse(ctx,
				finalPassInput,
				nil,
				true,
				resultCh,
				errCh)

			select {
			case result := <-resultCh:
				parseResults[0] = result.stack
			case err := <-errCh:
				cancel()
				return nil, err
			}
		}
	} else {
		// Loop until we have a single reduced stack
		p.concurrency--
		for workers := make([]*parserWorker, p.concurrency); p.concurrency >= 1; p.concurrency-- {
			// TODO: Fill the right info
			for i := 0; i < p.concurrency; i++ {
				stackLeft := parseResults[i]
				stackRight := parseResults[i+1]

				stack := stackLeft.Combine()
				input := CombineLOS(tokens, stackRight)

				workers[i] = &parserWorker{
					parser:    p,
					id:        i,
					stack:     stack,
					newNTList: NewListOfStacks[Token](pools[i]),
				}

				go workers[i].parse(ctx, input, nil, true, resultCh, errCh)
			}

			completed = 0
			for completed < p.concurrency {
				select {
				case result := <-resultCh:
					parseResults[result.threadNum] = result.stack
					completed++
				case err := <-errCh:
					cancel()
					return nil, err
				}
			}
		}
	}

	// Pop tokens until the last non-terminal is found.
	// TODO: Check if this makes sense, the former approach looked for the first one
	// TODO: but it wasn't working for AOPP.
	var root *Token
	for token := parseResults[0].Pop(); token != nil; token = parseResults[0].Pop() {
		if !token.Type.IsTerminal() {
			root = token
		}
	}

	return root, nil
}

type parserWorker struct {
	parser *Parser

	id int

	newNTList *ListOfStacks[Token]
	stack     *listOfTokenPointerStacks
}

type parseResult struct {
	threadNum int
	stack     *listOfTokenPointerStacks
}

func (w *parserWorker) parse(ctx context.Context, tokens *ListOfStacks[Token], nextToken *Token, finalPass bool, resultCh chan<- parseResult, errCh chan<- error) {
	tokensIt := tokens.HeadIterator()

	// newNonTerminalsList := NewListOfStacks[Token](w.stackPool)
	// stack := newListOfTokenPointerStacks(w.ptrStackPool)

	// If the thread is the first, push a # onto the stack
	// Otherwise, push the first inputToken onto the stack
	if !finalPass {
		if w.id == 0 {
			w.stack.Push(&Token{
				Type:  TokenTerm,
				Value: nil,
				// Lexeme:     "",
				Precedence: PrecEmpty,
				Next:       nil,
				Child:      nil,
			})
		} else {
			t := tokensIt.Next()
			t.Precedence = PrecEmpty
			w.stack.Push(t)
			// precToken.Precedence = PrecEmpty
			// stack.Push(precToken)
		}

		// If the thread is the last, push a # onto the tokens m
		// Otherwise, push onto the tokens m the first inputToken of the next tokens m
		if w.id == w.parser.concurrency-1 {
			tokens.Push(Token{
				Type:  TokenTerm,
				Value: nil,
				// Lexeme:     "",
				Precedence: PrecEmpty,
				Next:       nil,
				Child:      nil,
			})
		} else if nextToken != nil {
			tokens.Push(*nextToken)
		}
	}

	var pos int
	var lhsToken *Token

	var rhs []TokenType
	var rhsTokens []*Token

	rhsBuf := make([]TokenType, w.parser.MaxRHSLength)
	rhsTokensBuf := make([]*Token, w.parser.MaxRHSLength)

	newNonTerm := &Token{
		Type:  TokenEmpty,
		Value: nil,
		// Lexeme:     "",
		Precedence: PrecEmpty,
		Next:       nil,
		Child:      nil,
	}

	// Iterate over the tokens
	// If this is the first worker, start reading from the input stack, otherwise begin with the last
	// token of the previous stack.
	for inputToken := tokensIt.Next(); inputToken != nil; {
		//If the current inputToken is a non-terminal, push it onto the stack with no precedence relation
		if !inputToken.Type.IsTerminal() {
			inputToken.Precedence = PrecEmpty
			w.stack.Push(inputToken)

			inputToken = tokensIt.Next()
			continue
		}

		//Find the first terminal on the stack and get the precedence between it and the current tokens inputToken
		firstTerminal := w.stack.FirstTerminal()

		var prec Precedence
		if firstTerminal == nil {
			prec = w.parser.precedence(TokenTerm, inputToken.Type)
		} else {
			prec = w.parser.precedence(firstTerminal.Type, inputToken.Type)
		}

		// If it's equal in precedence or yields, push the inputToken onto the stack with its precedence relation.
		if prec == PrecEquals || prec == PrecYields {
			inputToken.Precedence = prec
			w.stack.Push(inputToken)

			inputToken = tokensIt.Next()
		} else if prec == PrecTakes || prec == PrecAssociative {
			//If there are no tokens yielding precedence on the stack, push inputToken onto the stack.
			//Otherwise, perform a reduction
			if w.stack.YieldingPrecedence() == 0 {
				inputToken.Precedence = prec
				w.stack.Push(inputToken)

				inputToken = tokensIt.Next()
			} else {
				pos = w.parser.MaxRHSLength - 1

				var token *Token
				// Pop tokens from the stack until one that yields precedence is reached, saving them in rhsBuf
				for token = w.stack.Pop(); token.Precedence != PrecYields && token.Precedence != PrecAssociative; token = w.stack.Pop() {
					rhsTokensBuf[pos] = token
					rhsBuf[pos] = token.Type
					pos--
				}
				rhsTokensBuf[pos] = token
				rhsBuf[pos] = token.Type

				//Pop one last token, if it's a non-terminal add it to rhsBuf, otherwise ignore it (push it again onto the stack)
				token = w.stack.Pop()
				if token.Type.IsTerminal() {
					w.stack.Push(token)
				} else {
					pos--
					rhsTokensBuf[pos] = token
					rhsBuf[pos] = token.Type

					w.stack.UpdateFirstTerminal()
				}

				//Obtain the actual rhs from the buffers
				rhsTokens = rhsTokensBuf[pos:]
				rhs = rhsBuf[pos:]

				//Find corresponding lhs and ruleNum
				lhs, ruleNum := w.parser.findMatch(rhs)
				if lhs == TokenEmpty {
					errCh <- fmt.Errorf("could not find match for rhs %v", rhs)
					return
				}

				newNonTerm.Type = lhs
				lhsToken = w.newNTList.Push(*newNonTerm)

				//Execute the semantic action
				w.parser.Func(ruleNum, lhsToken, rhsTokens, w.id)

				//Push the new nonterminal onto the stack
				w.stack.Push(lhsToken)
			}
		} else {
			//If there's no precedence relation, abort the parsing
			errCh <- fmt.Errorf("no precedence relation found")
			return
		}
	}

	resultCh <- parseResult{w.id, w.stack}
}

func (p *Parser) precedence(t1 TokenType, t2 TokenType) Precedence {
	v1 := t1.Value()
	v2 := t2.Value()

	flatElementPos := v1*p.NumTerminals + v2
	elem := p.BitPackedPrecedenceMatrix[flatElementPos/32]
	pos := uint((flatElementPos % 32) * 2)

	return Precedence((elem >> pos) & 0x3)
}

func (p *Parser) findMatch(rhs []TokenType) (TokenType, uint16) {
	var pos uint16

	for _, key := range rhs {
		//Skip the value and rule num for each node (except the last)
		pos += 2
		numIndices := p.CompressedRules[pos]
		if numIndices == 0 {
			return TokenEmpty, 0
		}

		pos++
		low := uint16(0)
		high := numIndices - 1
		startPos := pos
		foundNext := false

		for low <= high {
			indexPos := low + (high-low)/2
			pos = startPos + indexPos*2
			curKey := p.CompressedRules[pos]

			if uint16(key) < curKey {
				high = indexPos - 1
			} else if uint16(key) > curKey {
				low = indexPos + 1
			} else {
				pos = p.CompressedRules[pos+1]
				foundNext = true
				break
			}
		}
		if !foundNext {
			return TokenEmpty, 0
		}
	}

	return TokenType(p.CompressedRules[pos]), p.CompressedRules[pos+1]
}
