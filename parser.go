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

		if thread < p.concurrency-1 {
			nextInputListIter := tokensLists[thread+1].HeadIterator()
			nextToken = nextInputListIter.Next()
		}

		workers[thread] = &parserWorker{
			parser: p,
			id:     thread,

			stackPool:    pools[thread],
			ptrStackPool: ptrPools[thread],
		}

		go workers[thread].parse(ctx,
			tokensLists[thread],
			nextToken,
			false,
			resultCh,
			errCh)
	}

	parseResults := make([]parseResult, p.concurrency)
	completed := 0

	for completed < p.concurrency {
		select {
		case result := <-resultCh:
			parseResults[result.threadNum] = result
			completed++
		case err := <-errCh:
			cancel()
			return nil, err
		}
	}

	//If the number of threads is greater than one, a final pass is required
	if p.concurrency > 1 {
		//Create the final input by joining together the stacks from the previous step
		finalPassInput := NewListOfStacks[Token](stackPoolFinalPass)

		for i := 0; i < p.concurrency; i++ {
			iterator := parseResults[i].stack.HeadIterator()

			//Ignore the first token
			iterator.Next()

			for token := iterator.Next(); token != nil; token = iterator.Next() {
				finalPassInput.Push(*token)
			}
		}

		workers[0].stackPool = stackPoolNewNonterminalsFinalPass
		workers[0].ptrStackPool = stackPtrPoolFinalPass

		p.concurrency = 1

		go workers[0].parse(ctx,
			finalPassInput,
			nil,
			true,
			resultCh,
			errCh)

		select {
		case result := <-resultCh:
			parseResults[0] = result
		case err := <-errCh:
			cancel()
			return nil, err
		}
	}

	//Pop tokens from the stack until a nonterminal is found
	token := parseResults[0].stack.Pop()
	for token.Type.IsTerminal() {
		token = parseResults[0].stack.Pop()
	}

	return token, nil
}

type parserWorker struct {
	parser *Parser

	id int

	stackPool    *Pool[stack[Token]]
	ptrStackPool *Pool[tokenPointerStack]
}

type parseResult struct {
	threadNum int
	stack     *listOfTokenPointerStacks
}

func (w *parserWorker) parse(ctx context.Context, tokens *ListOfStacks[Token], nextToken *Token, finalPass bool, resultCh chan<- parseResult, errCh chan<- error) {
	tokensIt := tokens.HeadIterator()

	newNonTerminalsList := NewListOfStacks[Token](w.stackPool)
	stack := newListOfTokenPointerStacks(w.ptrStackPool)

	// If the thread is the first, push a # onto the stack
	// Otherwise, push the first inputToken onto the stack
	if w.id == 0 {
		stack.Push(&Token{
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

		stack.Push(t)
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
	} else {
		tokens.Push(*nextToken)
	}

	numYieldsPrec := 0

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

	i := 0
	// Iterate over the tokens m
	for inputToken := tokensIt.Next(); inputToken != nil; i++ {
		//If the current inputToken is a non-terminal, push it onto the stack with no precedence relation
		if !inputToken.Type.IsTerminal() {
			inputToken.Precedence = PrecEmpty
			stack.Push(inputToken)

			inputToken = tokensIt.Next()
			continue
		}

		//Find the first terminal on the stack and get the precedence between it and the current tokens inputToken
		firstTerminal := stack.FirstTerminal()
		prec := w.parser.precedence(firstTerminal.Type, inputToken.Type)
		switch prec {
		// If it yields precedence, push the tokens inputToken onto the stack with that precedence relation.
		// Also increment the counter of the number of tokens yielding precedence.
		case PrecYields:
			inputToken.Precedence = PrecYields
			stack.Push(inputToken)
			numYieldsPrec++

			inputToken = tokensIt.Next()
		// If it's equal in precedence, push the tokens inputToken onto the stack with that precedence relation
		case PrecEquals:
			inputToken.Precedence = PrecEquals
			stack.Push(inputToken)

			inputToken = tokensIt.Next()
		// If it takes precedence, the next action depends on whether there are tokens that yield precedence onto the stack.
		case PrecTakes:
			//If there are no tokens yielding precedence on the stack, push the tokens inputToken onto the stack with take precedence as precedence relation
			//Otherwise, perform a reduction
			if numYieldsPrec == 0 {
				inputToken.Precedence = PrecTakes
				stack.Push(inputToken)

				inputToken = tokensIt.Next()
			} else {
				pos = w.parser.MaxRHSLength - 1

				var token *Token
				// Pop tokens from the stack until one that yields precedence is reached, saving them in rhsBuf
				for token = stack.Pop(); token.Precedence != PrecYields; token = stack.Pop() {
					rhsTokensBuf[pos] = token
					rhsBuf[pos] = token.Type
					pos--
				}
				rhsTokensBuf[pos] = token
				rhsBuf[pos] = token.Type

				//Pop one last token, if it's a non-terminal add it to rhsBuf, otherwise ignore it (push it again onto the stack)
				token = stack.Pop()
				if token.Type.IsTerminal() {
					stack.Push(token)
				} else {
					pos--
					rhsTokensBuf[pos] = token
					rhsBuf[pos] = token.Type

					stack.UpdateFirstTerminal()
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

				//Push the new nonterminal onto the appropriate m to save it
				newNonTerm.Type = lhs
				lhsToken = newNonTerminalsList.Push(*newNonTerm)

				//Execute the semantic action
				w.parser.Func(ruleNum, lhsToken, rhsTokens, w.id)

				//Push the new nonterminal onto the stack
				stack.Push(lhsToken)

				//Decrement the counter of the number of tokens yielding precedence
				numYieldsPrec--
			}
		//If there's no precedence relation, abort the parsing
		default:
			errCh <- fmt.Errorf("no precedence relation found")
			return
		}
	}

	resultCh <- parseResult{w.id, stack}
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
