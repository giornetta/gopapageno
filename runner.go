package gopapageno

import (
	"context"
	"fmt"
	"io"
	"log"
	"runtime/pprof"
)

type Runner struct {
	Lexer  *Lexer
	Parser *Grammar

	concurrency        int
	initialConcurrency int
	reductionStrategy  ReductionStrategy

	avgTokenLength int

	logger *log.Logger

	cpuProfileWriter io.Writer
	memProfileWriter io.Writer
}

type RunnerOpt func(p *Runner)

func WithConcurrency(n int) RunnerOpt {
	return func(r *Runner) {
		if n <= 0 {
			n = 1
		}

		r.initialConcurrency = n
	}
}

func WithLogging(logger *log.Logger) RunnerOpt {
	return func(r *Runner) {
		if logger == nil {
			logger = discardLogger
		}

		r.logger = logger
	}
}

func WithCPUProfiling(w io.Writer) RunnerOpt {
	return func(r *Runner) {
		r.cpuProfileWriter = w
	}
}

func WithMemoryProfiling(w io.Writer) RunnerOpt {
	return func(r *Runner) {
		r.memProfileWriter = w
	}
}

func WithReductionStrategy(strat ReductionStrategy) RunnerOpt {
	return func(r *Runner) {
		r.reductionStrategy = strat
	}
}

const DefaultAverageTokenLength int = 4

func WithAverageTokenLength(length int) RunnerOpt {
	return func(r *Runner) {
		r.avgTokenLength = length
	}
}

func NewRunner(lexer *Lexer, parser *Grammar, opts ...RunnerOpt) *Runner {
	r := &Runner{
		Lexer:  lexer,
		Parser: parser,

		concurrency:        1,
		initialConcurrency: 1,
		reductionStrategy:  ReductionSweep,
		avgTokenLength:     DefaultAverageTokenLength,
		logger:             discardLogger,
		cpuProfileWriter:   nil,
		memProfileWriter:   nil,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *Runner) Run(ctx context.Context, src []byte) (*Token, error) {
	r.concurrency = r.initialConcurrency

	// Profiling
	cleanupFunc := r.startProfiling()
	defer cleanupFunc()

	// Run preamble functions before anything else.
	if r.Lexer.PreambleFunc != nil {
		r.Lexer.PreambleFunc(len(src), r.concurrency)
	}

	if r.Parser.PreambleFunc != nil {
		r.Parser.PreambleFunc(len(src), r.concurrency)
	}

	// Initialize Scanner and Grammar
	scanner := r.Lexer.Scanner(src, r.concurrency, r.avgTokenLength)
	parser := r.Parser.Parser(src, r.concurrency, r.avgTokenLength, r.reductionStrategy)

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
	if len(tokensLists) < r.concurrency {
		r.concurrency = len(tokensLists)
		r.logger.Printf("Not enough stacks in lexer output, lowering parser concurrency to %d", r.concurrency)
	}

	token, err := parser.Parse(ctx, tokensLists)
	if err != nil {
		return nil, fmt.Errorf("could not parse: %w", err)
	}

	return token, nil
}

func (r *Runner) startProfiling() func() {
	if r.cpuProfileWriter == nil || r.cpuProfileWriter != io.Discard {
		return func() {}
	}

	if err := pprof.StartCPUProfile(r.cpuProfileWriter); err != nil {
		log.Printf("could not start CPU profiling: %v", err)
	}

	return func() {
		if r.memProfileWriter != nil && r.memProfileWriter != io.Discard {
			if err := pprof.WriteHeapProfile(r.memProfileWriter); err != nil {
				log.Printf("Could not write memory profile: %v", err)
			}
		}

		pprof.StopCPUProfile()
	}
}
