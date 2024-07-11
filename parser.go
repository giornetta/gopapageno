package gopapageno

import (
	"context"
	"fmt"
	"io"
	"log"
	"runtime/pprof"
)

var (
	discardLogger = log.New(io.Discard, "", 0)
)

type Stacker interface {
	HeadIterator() *ParserStackIterator
	Combine() Stacker
	CombineLOS(pool *Pool[stack[Token]]) *ListOfStacks[Token]
	LastNonterminal() (*Token, error)
}

type StackerIterator interface {
}

type parseResult struct {
	threadNum int
	stack     Stacker
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

	concurrency        int
	initialConcurrency int
	reductionStrategy  ReductionStrategy

	avgTokenLength int

	logger *log.Logger

	cpuProfileWriter io.Writer
	memProfileWriter io.Writer

	pools parserPools
}

type parserPools struct {
	stacks       []*Pool[stack[*Token]]
	nonterminals []*Pool[Token]

	// These are only used when parsing using COPP.
	stateStacks []*Pool[stack[CyclicAutomataState]]

	// These are only used when reducing using a single sweep.
	sweepInput *Pool[stack[Token]]
	sweepStack *Pool[stack[*Token]]

	sweepStateStack *Pool[stack[CyclicAutomataState]]
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

		p.initialConcurrency = n
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

const DefaultAverageTokenLength int = 4

func WithAverageTokenLength(length int) ParserOpt {
	return func(p *Parser) {
		p.avgTokenLength = length
	}
}

func NewParser(
	lexer *Lexer,
	numTerminals uint16, numNonterminals uint16, maxRHSLength int,
	rules []Rule, compressedRules []uint16,
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
		PrecedenceMatrix:          precedenceMatrix,
		BitPackedPrecedenceMatrix: bitPackedPrecedenceMatrix,
		Func:                      fn,
		ParsingStrategy:           strategy,
		concurrency:               1,
		initialConcurrency:        1,
		reductionStrategy:         ReductionSweep,
		avgTokenLength:            DefaultAverageTokenLength,
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
	p.concurrency = p.initialConcurrency

	// Profiling
	cleanupFunc := p.startProfiling()
	defer cleanupFunc()

	// Run Prealloc Functions
	if p.Lexer.PreallocFunc != nil {
		p.Lexer.PreallocFunc(len(src), p.concurrency)
	}

	if p.PreallocFunc != nil {
		p.PreallocFunc(len(src), p.concurrency)
	}

	// Initialize Scanner
	scanner := p.Lexer.Scanner(src, p.concurrency, p.avgTokenLength)

	// Allocate
	p.init(src)

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
	parseResults := make([]Stacker, p.concurrency)
	workers := make([]*parserWorker, p.concurrency)

	// First parallel pass of the algorithm.
	for thread := 0; thread < p.concurrency; thread++ {
		var nextToken *Token

		// If the thread is not the last, also take the first token of the next stack for lookahead.
		if thread < p.concurrency-1 {
			nextInputListIter := tokensLists[thread+1].HeadIterator()
			nextToken = nextInputListIter.Next()
		}

		workers[thread] = &parserWorker{
			parser: p,
			id:     thread,
			ntPool: p.pools.nonterminals[thread],
		}

		var s Stacker
		if p.ParsingStrategy != COPP {
			s = NewParserStack(p.pools.stacks[thread])
		} else {
			s = NewCyclicParserStack(p.pools.stacks[thread], p.pools.stateStacks[thread])
		}

		go workers[thread].parse(ctx, s, tokensLists[thread], nextToken, false, resultCh, errCh)
	}

	if err := collectResults(parseResults, resultCh, errCh, p.concurrency); err != nil {
		cancel()
		return nil, err
	}

	//If the number of threads is greater than one, results must be combined and work should continue.
	reductionPasses := 0

	// Loop until we have a single reduced stack
	for p.concurrency--; p.concurrency >= 1; p.concurrency-- {
		if p.reductionStrategy == ReductionSweep || (p.reductionStrategy == ReductionMixed && reductionPasses >= 2) {

			// Nullifies the previous p.concurrency--
			p.concurrency++

			// Create the final input by joining together the stacks from the previous step.
			stack := parseResults[0].Combine()
			input := p.CombineSweepLOS(p.pools.sweepInput, parseResults[1:])

			// Sets correct concurrency level for final sweep.
			p.concurrency = 1

			go workers[0].parse(ctx, stack, input, nil, true, resultCh, errCh)

			if err := collectResults(parseResults, resultCh, errCh, 1); err != nil {
				cancel()
				return nil, err
			}
		} else {
			for i := 0; i < p.concurrency; i++ {
				stackLeft := parseResults[i]
				stackRight := parseResults[i+1]

				// TODO: Fix CombineNoAlloc
				stack := stackLeft.Combine()

				// TODO: I should find a way to make this work without creating a new LOS for the inputs.
				// Unfortunately the new stack depends on the content of tokensLists[i] since its elements are stored there.
				// We can't erase the old input easily to reuse its storage.
				// TODO: Maybe allocate 2 * c LOS so that we can alternate?
				input := stackRight.CombineLOS(tokensLists[i].pool)

				go workers[i].parse(ctx, stack, input, nil, true, resultCh, errCh)
			}

			if err := collectResults(parseResults, resultCh, errCh, p.concurrency); err != nil {
				cancel()
				return nil, err
			}

			reductionPasses++
		}

	}

	// Pop tokens until a non-terminal is found.
	return parseResults[0].LastNonterminal()
}

func (p *Parser) init(src []byte) {
	srcLen := len(src)

	stackPoolBaseSize := stacksCount[*Token](src, p.concurrency, p.avgTokenLength)
	ntPoolBaseSize := srcLen / p.avgTokenLength / p.concurrency

	// Initialize memory pools for stacks.
	p.pools.stacks = make([]*Pool[stack[*Token]], p.concurrency)

	// Initialize memory pools for cyclic states.
	if p.ParsingStrategy == COPP {
		p.pools.stateStacks = make([]*Pool[stack[CyclicAutomataState]], p.concurrency)
	}

	// Initialize pools to hold pointers to tokens generated by the reduction steps.
	p.pools.nonterminals = make([]*Pool[Token], p.concurrency)

	for thread := 0; thread < p.concurrency; thread++ {
		// TODO: Does this need more work?
		stackPoolMultiplier := .25
		if p.reductionStrategy == ReductionParallel {
			//stackPoolMultiplier = p.concurrency - thread
		}

		p.pools.stacks[thread] = NewPool[stack[*Token]](int(float64(stackPoolBaseSize)*stackPoolMultiplier), WithConstructor[stack[*Token]](newStack[*Token]))

		if p.ParsingStrategy == COPP {
			p.pools.stateStacks[thread] = NewPool[stack[CyclicAutomataState]](int(float64(stackPoolBaseSize)*stackPoolMultiplier), WithConstructor[stack[CyclicAutomataState]](newStack[CyclicAutomataState]))
		}

		p.pools.nonterminals[thread] = NewPool[Token](ntPoolBaseSize)
	}

	// TODO: Remove or change this part to reflect the correct sweep reductionStrategy.
	if p.reductionStrategy == ReductionSweep || p.reductionStrategy == ReductionMixed {
		inputPoolBaseSize := stacksCount[Token](src, p.concurrency, p.avgTokenLength)

		p.pools.sweepInput = NewPool[stack[Token]](inputPoolBaseSize, WithConstructor[stack[Token]](newStack[Token]))
		p.pools.sweepStack = NewPool[stack[*Token]](stackPoolBaseSize, WithConstructor[stack[*Token]](newStack[*Token]))

		if p.ParsingStrategy == COPP {
			p.pools.sweepStateStack = NewPool[stack[CyclicAutomataState]](stackPoolBaseSize, WithConstructor[stack[CyclicAutomataState]](newStack[CyclicAutomataState]))
		}
	}
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

func (p *Parser) startProfiling() func() {
	if p.cpuProfileWriter == nil || p.cpuProfileWriter != io.Discard {
		return func() {}
	}

	if err := pprof.StartCPUProfile(p.cpuProfileWriter); err != nil {
		log.Printf("could not start CPU profiling: %v", err)
	}

	return func() {
		if p.memProfileWriter != nil && p.memProfileWriter != io.Discard {
			if err := pprof.WriteHeapProfile(p.memProfileWriter); err != nil {
				log.Printf("Could not write memory profile: %v", err)
			}
		}

		pprof.StopCPUProfile()
	}
}

func collectResults(results []Stacker, resultCh <-chan parseResult, errCh <-chan error, n int) error {
	completed := 0
	for completed < n {
		select {
		case result := <-resultCh:
			results[result.threadNum] = result.stack
			completed++
		case err := <-errCh:
			return err
		}
	}

	return nil
}
