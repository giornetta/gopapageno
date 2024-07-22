package gopapageno

import (
	"context"
	"fmt"
	"io"
	"log"
	"runtime/pprof"
)

type Runner[T Tokener] struct {
	Lexer   *Lexer[T]
	Grammar *Grammar[T]

	concurrency       int
	reductionStrategy ReductionStrategy

	avgTokenLength int

	logger *log.Logger

	cpuProfileWriter io.Writer
	memProfileWriter io.Writer
}

func NewRunner[T Tokener](l *Lexer[T], g *Grammar[T], opts ...RunnerOpt[T]) *Runner[T] {
	r := &Runner[T]{
		Lexer:          l,
		Grammar:        g,
		avgTokenLength: DefaultAverageTokenLength,
		logger:         discardLogger,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

type RunnerOpt[T Tokener] func(r *Runner[T])

func WithConcurrency[T Tokener](n int) RunnerOpt[T] {
	return func(r *Runner[T]) {
		if n <= 0 {
			n = 1
		}

		r.concurrency = n
	}
}

func WithLogging[T Tokener](logger *log.Logger) RunnerOpt[T] {
	return func(r *Runner[T]) {
		if logger == nil {
			logger = discardLogger
		}

		r.logger = logger
	}
}

func WithCPUProfiling[T Tokener](w io.Writer) RunnerOpt[T] {
	return func(r *Runner[T]) {
		r.cpuProfileWriter = w
	}
}

func WithMemoryProfiling[T Tokener](w io.Writer) RunnerOpt[T] {
	return func(r *Runner[T]) {
		r.memProfileWriter = w
	}
}

func WithReductionStrategy[T Tokener](strat ReductionStrategy) RunnerOpt[T] {
	return func(r *Runner[T]) {
		r.reductionStrategy = strat
	}
}

const DefaultAverageTokenLength int = 4

func WithAverageTokenLength[T Tokener](length int) RunnerOpt[T] {
	return func(r *Runner[T]) {
		r.avgTokenLength = length
	}
}

func (r *Runner[T]) Run(ctx context.Context, src []byte) (*T, error) {
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

func (r *Runner[T]) startProfiling() func() {
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

type Parser[T Tokener] interface {
	Parse(ctx context.Context, tokenLists []*LOS[T]) (*T, error)
}
