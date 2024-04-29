package gopapageno

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"runtime/pprof"
	"slices"
)

var (
	discardLogger = log.New(io.Discard, "", 0)
)

type Rule struct {
	Lhs TokenType
	Rhs []TokenType
}

type ParserFunc func(rule uint16, lhs *Token, rhs []*Token, thread int)

// A ReductionStrategy defines which kind of algorithm should be executed
// when collecting and running multiple parsing passes.
type ReductionStrategy uint8

const (
	// ReductionSweep will run a single serial pass after combining data from the first `n` parallel runs.
	ReductionSweep ReductionStrategy = iota
	// ReductionParallel will combine adjacent parsing results and recursively run `n-1` parallel runs until one stack remains.
	ReductionParallel
	// ReductionMixed will run a limited number of parallel passes, then combine the remaining inputs to perform a final serial pass.
	ReductionMixed
)

type ParsingStrategy uint8

const (
	// OPP is Operator-Precedence Parsing. It is the original parsing reductionStrategy.
	OPP ParsingStrategy = iota
	// AOPP is Associative Operator Precedence Parsing.
	AOPP
	// COPP is Cyclic Operator Precedence Parsing.
	COPP
)

func (s ParsingStrategy) String() string {
	switch s {
	case OPP:
		return "OPP"
	case AOPP:
		return "AOPP"
	case COPP:
		return "COPP"
	default:
		return "UNKNOWN"
	}
}

type Parser struct {
	Lexer *Lexer

	NumTerminals    uint16
	NumNonterminals uint16

	MaxRHSLength    int
	Rules           []Rule
	CompressedRules []uint16

	Prefixes        [][]TokenType
	MaxPrefixLength int

	PrecedenceMatrix          [][]Precedence
	BitPackedPrecedenceMatrix []uint64

	Func ParserFunc

	PreallocFunc PreallocFunc

	ParsingStrategy ParsingStrategy

	concurrency       int
	reductionStrategy ReductionStrategy

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

func WithReductionStrategy(strat ReductionStrategy) ParserOpt {
	return func(p *Parser) {
		p.reductionStrategy = strat
	}
}

func NewParser(
	lexer *Lexer,
	numTerminals uint16, numNonterminals uint16, maxRHSLength int,
	rules []Rule, compressedRules []uint16,
	prefixes [][]TokenType, maxPrefixLength int,
	precedenceMatrix [][]Precedence, bitPackedPrecedenceMatrix []uint64,
	fn ParserFunc,
	strategy ParsingStrategy,
	opts ...ParserOpt,
) *Parser {
	parser := &Parser{
		Lexer:                     lexer,
		NumTerminals:              numTerminals,
		NumNonterminals:           numNonterminals,
		MaxRHSLength:              maxRHSLength,
		Rules:                     rules,
		CompressedRules:           compressedRules,
		Prefixes:                  prefixes,
		MaxPrefixLength:           maxPrefixLength,
		PrecedenceMatrix:          precedenceMatrix,
		BitPackedPrecedenceMatrix: bitPackedPrecedenceMatrix,
		Func:                      fn,
		ParsingStrategy:           strategy,
		concurrency:               1,
		reductionStrategy:         ReductionSweep,
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

	// TODO: Where does this number come from?
	avgCharsPerToken := 4.0

	stackPoolBaseSize := math.Ceil(float64(srcLen) / avgCharsPerToken / float64(stackSize) / float64(p.concurrency))
	stackPtrPoolBaseSize := math.Ceil(float64(srcLen) / avgCharsPerToken / float64(pointerStackSize) / float64(p.concurrency))
	ntPoolBaseSize := math.Ceil(float64(srcLen) / avgCharsPerToken / float64(p.concurrency))

	// Initialize memory pools for input lists.
	pools := make([]*Pool[stack[Token]], p.concurrency)

	// Initialize memory pools for stacks.
	ptrPools := make([]*Pool[stack[*Token]], p.concurrency)

	// Initialize pools to hold pointers to tokens generated by the reduction steps.
	ntPools := make([]*Pool[Token], p.concurrency)

	for thread := 0; thread < p.concurrency; thread++ {
		pools[thread] = NewPool[stack[Token]](int(stackPoolBaseSize*0.8), WithConstructor[stack[Token]](newStack[Token]))
		ptrPools[thread] = NewPool[stack[*Token]](int(stackPtrPoolBaseSize), WithConstructor[stack[*Token]](newStack[*Token]))
		ntPools[thread] = NewPool[Token](int(ntPoolBaseSize))
	}

	// TODO: Remove or change this part to reflect the correct sweep reductionStrategy.
	stackPoolFinalPass := NewPool[stack[Token]](int(math.Ceil(stackPoolBaseSize*0.1*float64(p.concurrency))), WithConstructor[stack[Token]](newStack[Token]))
	stackPtrPoolFinalPass := NewPool[stack[*Token]](int(math.Ceil(stackPtrPoolBaseSize*0.1)), WithConstructor[stack[*Token]](newStack[*Token]))

	// TODO: Investigate this section better.
	// Old code forced a GC Run to occur, so that it would - hopefully - stop GCs from happening again during computation.
	// However a GC run can still be very slow.
	// runtime.GC()

	// This new version stops the GC from running entirely.
	// debug.SetGCPercent(-1)

	// Deferring this will cause the GC to still run at the end of computation...
	// defer debug.SetGCPercent(1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tokensLists, err := scanner.Lex(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not lex: %w", err)
	}
	// If there are not enough stacks in the input, reduce the number of threads.
	// The input is split by splitting stacks, not stack contents.
	if len(tokensLists) < p.concurrency {
		p.concurrency = len(tokensLists)
		p.logger.Printf("Not enough stacks in lexer output, lowering parser concurrency to %d", p.concurrency)
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

			stack:  NewParserStack(ptrPools[thread]),
			ntPool: ntPools[thread],
		}

		if p.ParsingStrategy != COPP {
			go workers[thread].parse(ctx, tokensLists[thread], nextToken, false, resultCh, errCh)
		} else {
			go workers[thread].parseCyclic(ctx, tokensLists[thread], nextToken, false, resultCh, errCh)
		}

	}

	parseResults := make([]*ParserStack, p.concurrency)
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
	//If the number of threads is greater than one, a final pass is required
	if p.concurrency > 1 {
		if p.reductionStrategy == ReductionSweep {
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

			workers[0].stack = NewParserStack(stackPtrPoolFinalPass)

			p.concurrency = 1

			if p.ParsingStrategy != COPP {
				go workers[0].parse(ctx, finalPassInput, nil, true, resultCh, errCh)
			} else {
				go workers[0].parseCyclic(ctx, finalPassInput, nil, true, resultCh, errCh)
			}

			select {
			case result := <-resultCh:
				parseResults[0] = result.stack
			case err := <-errCh:
				cancel()
				return nil, err
			}
		} else if p.reductionStrategy == ReductionParallel {
			// Loop until we have a single reduced stack
			for p.concurrency--; p.concurrency >= 1; p.concurrency-- {
				for i := 0; i < p.concurrency; i++ {
					stackLeft := parseResults[i]
					stackRight := parseResults[i+1]

					// TODO: Fix CombineNoAlloc
					stack := stackLeft.Combine()
					// stackLeft.CombineNoAlloc()

					// TODO: I should find a way to make this work without creating a new LOS for the inputs.
					// Unfortunately the new stack depends on the content of tokensLists[i] since its elements are stored there.
					// We can't erase the old input easily to reuse its storage.
					// TODO: Maybe allocate 2 * c LOS so that we can alternate?
					input := CombineLOS(tokensLists[i], stackRight)

					workers[i].stack = stack

					if p.ParsingStrategy != COPP {
						go workers[i].parse(ctx, input, nil, true, resultCh, errCh)
					} else {
						go workers[i].parseCyclic(ctx, input, nil, true, resultCh, errCh)
					}
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
	}

	// Pop tokens until a non-terminal is found.
	for token := parseResults[0].Pop(); token != nil; token = parseResults[0].Pop() {
		if !token.Type.IsTerminal() {
			return token, nil
		}
	}

	return nil, fmt.Errorf("no non-terminal token found after parsing")
}

type parserWorker struct {
	parser *Parser

	id int

	ntPool *Pool[Token]
	stack  *ParserStack
}

type parseResult struct {
	threadNum int
	stack     *ParserStack
}

func (w *parserWorker) parse(ctx context.Context, tokens *ListOfStacks[Token], nextToken *Token, finalPass bool, resultCh chan<- parseResult, errCh chan<- error) {
	tokensIt := tokens.HeadIterator()

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

	newNonTerm := Token{
		Type:       TokenEmpty,
		Value:      nil,
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
				// lhsToken = w.newNTList.Push(*newNonTerm)
				lhsToken = w.ntPool.Get()
				*lhsToken = newNonTerm

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

func (w *parserWorker) parseCyclic(ctx context.Context, tokens *ListOfStacks[Token], nextToken *Token, finalPass bool, resultCh chan<- parseResult, errCh chan<- error) {
	tokensIt := tokens.HeadIterator()

	// If the thread is the first, push a # onto the stack
	// Otherwise, push the first inputToken onto the stack
	if !finalPass {
		if w.id == 0 {
			w.stack.Push(&Token{
				Type:       TokenTerm,
				Value:      nil,
				Precedence: PrecEmpty,
				Next:       nil,
				Child:      nil,
			})
		} else {
			t := tokensIt.Next()
			t.Precedence = PrecEmpty
			w.stack.Push(t)
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

	newNonTerm := Token{
		Type:       TokenEmpty,
		Value:      nil,
		Precedence: PrecEmpty,
		Next:       nil,
		Child:      nil,
	}

	curRhsPrefix := make([]*Token, w.parser.MaxPrefixLength+1)
	curRhsPrefixTokens := make([]TokenType, w.parser.MaxPrefixLength+1)
	curRhsPrefixLen := 0

	prevRhsPrefix := make([]*Token, w.parser.MaxPrefixLength+1)
	prevRhsPrefixTokens := make([]TokenType, w.parser.MaxPrefixLength+1)
	prevRhsPrefixLen := 0

	// TODO: Maybe a stack for prefixes is better?
	// lastRhsPrefix := make([]*Token, w.parser.MaxPrefixLength+1)
	// lastRhsPrefixTokens := make([]TokenType, w.parser.MaxPrefixLength+1)
	// lastRhsPrefixLen := 0

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

		// If it yields precedence, PUSH the inputToken onto the stack with its precedence relation.
		if prec == PrecYields {
			inputToken.Precedence = prec
			t := w.stack.Push(inputToken)

			// If the current construction is a single nonterminal.
			if curRhsPrefixLen == 1 && !curRhsPrefix[0].Type.IsTerminal() {
				// Append input character to the current construction.
				curRhsPrefix[curRhsPrefixLen] = inputToken
				curRhsPrefixTokens[curRhsPrefixLen] = inputToken.Type
				curRhsPrefixLen++
			} else {
				// Otherwise, swap.
				copy(prevRhsPrefix, curRhsPrefix)
				copy(prevRhsPrefixTokens, curRhsPrefixTokens)
				prevRhsPrefixLen = curRhsPrefixLen

				for i := 1; i < curRhsPrefixLen; i++ {
					curRhsPrefix[i] = nil
					curRhsPrefixTokens[i] = TokenEmpty
				}
				curRhsPrefix[0] = t
				curRhsPrefixTokens[0] = t.Type
				curRhsPrefixLen = 1
			}

			inputToken = tokensIt.Next()
		} else if prec == PrecEquals {
			// If it is equals, it is probably a shift transition?

			// If the current construction is a single nonterminal.
			if curRhsPrefixLen == 1 && !curRhsPrefix[0].Type.IsTerminal() {
				// Prepend previous construction to current one; leaving the previous one untouched.
				copy(curRhsPrefix[prevRhsPrefixLen:], curRhsPrefix)
				copy(curRhsPrefix[:prevRhsPrefixLen], prevRhsPrefix)

				copy(curRhsPrefixTokens[prevRhsPrefixLen:], curRhsPrefixTokens)
				copy(curRhsPrefixTokens[:prevRhsPrefixLen], prevRhsPrefixTokens)

				curRhsPrefixLen += prevRhsPrefixLen
			}

			// Append input character to the current construction.
			curRhsPrefix[curRhsPrefixLen] = inputToken
			curRhsPrefixTokens[curRhsPrefixLen] = inputToken.Type
			curRhsPrefixLen++

			// If the construction has a suffix which is a double occurrence of a string produced by a Kleene-+.
			for _, prefix := range w.parser.Prefixes {
				if slices.Equal(prefix, curRhsPrefixTokens[:curRhsPrefixLen]) {
					idx := findRepeatedSuffixIndex(curRhsPrefixTokens[:curRhsPrefixLen])
					if idx != -1 {
						for i := idx; i < len(curRhsPrefixTokens); i++ {
							curRhsPrefix[i] = nil
							curRhsPrefixTokens[i] = TokenEmpty
						}
						curRhsPrefixLen = idx
						break
					}
				}
			}

			inputToken = tokensIt.Next()
		} else if prec == PrecTakes || prec == PrecAssociative {
			//If there are no tokens yielding precedence on the stack, push inputToken onto the stack.
			//Otherwise, perform a reduction. (Reduction == Pop/Shift move?)
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
				curRhsPrefixLen--
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
				// lhsToken = w.newNTList.Push(*newNonTerm)
				lhsToken = w.ntPool.Get()
				*lhsToken = newNonTerm

				//Execute the semantic action
				w.parser.Func(ruleNum, lhsToken, rhsTokens, w.id)

				//Push the new nonterminal onto the stack
				// w.stack.Push(lhsToken)

				// Push the new nonterminal onto the current construction.
				curRhsPrefix[curRhsPrefixLen] = lhsToken
				curRhsPrefixTokens[curRhsPrefixLen] = lhsToken.Type
				curRhsPrefixLen++
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

func findRepeatedSuffixIndex(seq []TokenType) int {
	n := len(seq)

	// Iterate over the sequence from the end
	for i := n - 1; i > 0; i-- {
		// Check if the suffix starting from index i is equal to the suffix of the sequence
		if slices.Equal(seq[i:], seq[:n-i]) {
			return i // Return the index of the second occurrence
		}
	}

	return -1 // No repeated suffix found
}
