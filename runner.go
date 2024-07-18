package gopapageno

import (
	"context"
	"fmt"
	"io"
	"log"
	"runtime/pprof"
)

type Runner struct {
	Lexer   *Lexer
	Grammar *Grammar

	concurrency       int
	reductionStrategy ReductionStrategy

	avgTokenLength int

	logger *log.Logger

	cpuProfileWriter io.Writer
	memProfileWriter io.Writer
}

func NewRunner(l *Lexer, g *Grammar, opts ...RunnerOpt) *Runner {
	r := &Runner{
		Lexer:          l,
		Grammar:        g,
		avgTokenLength: DefaultAverageTokenLength,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

type RunnerOpt func(r *Runner)

func WithConcurrency(n int) RunnerOpt {
	return func(r *Runner) {
		if n <= 0 {
			n = 1
		}

		r.concurrency = n
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

func (r *Runner) Run(ctx context.Context, src []byte) (*Token, error) {
	// Profiling
	cleanupFunc := r.startProfiling()
	defer cleanupFunc()

	sourceLen := len(src)

	// Run Preamble Functions
	if r.Lexer.PreambleFunc != nil {
		r.Lexer.PreambleFunc(sourceLen, r.concurrency)
	}

	if r.Grammar.PreambleFunc != nil {
		r.Grammar.PreambleFunc(sourceLen, r.concurrency)
	}

	scanner := r.Lexer.Scanner(src, r.concurrency, r.avgTokenLength)
	parser := r.Grammar.Parser(src, r.concurrency, r.avgTokenLength, r.reductionStrategy)

	// TODO: Should all code before this be moved into NewRunner?

	// Run the actual stuff.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tokensLists, err := scanner.Lex(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not lex: %w", err)
	}
	// If there are not enough stacks in the input, reduce the number of threads.
	// The input is split by splitting stacks, not stack contents.
	if len(tokensLists) < r.concurrency {
		r.logger.Printf("Not enough stacks in lexer output, lowering grammar concurrency to %d", r.concurrency)
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
