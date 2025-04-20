package benchmark

import (
	"context"
	"fmt"
	"math"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/giornetta/gopapageno"
)

type Entry[T any] struct {
	Filename       string
	ParallelFactor float64
	AvgTokenLength int
	Result         T
}

func Runner[T any](b *testing.B, parsingStrategy gopapageno.ParsingStrategy, reductionStrategy gopapageno.ReductionStrategy, newLexer func() *gopapageno.Lexer, newGrammar func() *gopapageno.Grammar, entries []*Entry[T]) {
	threads := int(math.Min(float64(runtime.NumCPU()), 32))

	b.Run(fmt.Sprintf("strategy=%s", parsingStrategy), func(b *testing.B) {
		for _, entry := range entries {
			b.Run(fmt.Sprintf("file=%s", path.Base(entry.Filename)), func(b *testing.B) {
				for c := 1; c <= threads; c++ {
					b.Run(fmt.Sprintf("goroutines=%d", c), func(b *testing.B) {
						b.Run(fmt.Sprintf("reduction=%s", reductionStrategy), func(b *testing.B) {
							bytes, err := os.ReadFile(entry.Filename)
							if err != nil {
								b.Fatalf("could not read source file %s: %v", entry.Filename, err)
							}

							b.SetBytes(0)

							r := gopapageno.NewRunner(
								newLexer(),
								newGrammar(),
								gopapageno.WithConcurrency(c),
								gopapageno.WithReductionStrategy(reductionStrategy),
								gopapageno.WithParallelFactor(entry.ParallelFactor),
								gopapageno.WithAverageTokenLength(entry.AvgTokenLength),
							)

							Run(b, r, bytes)
						})
					})
				}
			})
		}
	})
}

func ParserRunner[T any](b *testing.B, parsingStrategy gopapageno.ParsingStrategy, reductionStrategy gopapageno.ReductionStrategy, newLexer func() *gopapageno.Lexer, newGrammar func() *gopapageno.Grammar, entries []*Entry[T]) {
	threads := int(math.Min(float64(runtime.NumCPU()), 32))

	b.Run(fmt.Sprintf("strategy=%s", parsingStrategy), func(b *testing.B) {
		for _, entry := range entries {
			b.Run(fmt.Sprintf("file=%s", path.Base(entry.Filename)), func(b *testing.B) {
				for c := 1; c <= threads; c++ {
					b.Run(fmt.Sprintf("goroutines=%d", c), func(b *testing.B) {
						b.Run(fmt.Sprintf("reduction=%s", reductionStrategy), func(b *testing.B) {
							bytes, err := os.ReadFile(entry.Filename)
							if err != nil {
								b.Fatalf("could not read source file %s: %v", entry.Filename, err)
							}

							r := gopapageno.NewRunner(
								newLexer(),
								newGrammar(),
								gopapageno.WithConcurrency(c),
								gopapageno.WithReductionStrategy(reductionStrategy),
								gopapageno.WithParallelFactor(entry.ParallelFactor),
								gopapageno.WithAverageTokenLength(entry.AvgTokenLength),
							)

							b.StopTimer()
							b.ResetTimer()
							b.StartTimer()

							for b.Loop() {
								_, err := runParser(b, r, bytes)
								if err != nil {
									b.Fatalf("could not parse source file: %v", err)
								}
							}
						})
					})
				}
			})
		}
	})
}

func RunExpect[T comparable](b *testing.B, r *gopapageno.Runner, bytes []byte, expected T) {
	b.StopTimer()
	b.ResetTimer()

	ctx := context.Background()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		result, err := r.Run(ctx, bytes)
		if err != nil {
			b.Fatalf("could not parse source file: %v", err)
		}

		if *result.Value.(*T) != expected {
			b.Fatalf("expected %v, got %v", expected, *result.Value.(*T))
		}
	}
}

func Run(b *testing.B, r *gopapageno.Runner, bytes []byte) {
	b.StopTimer()
	b.ResetTimer()

	ctx := context.Background()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := r.Run(ctx, bytes)
		if err != nil {
			b.Fatalf("could not parse source file: %v", err)
		}
	}
}

func runParser(b *testing.B, r *gopapageno.Runner, src []byte) (*gopapageno.Token, error) {
	b.StopTimer()

	r.Options.Concurrency = r.Options.InitialConcurrency

	// Run preamble functions before anything else.
	if r.Lexer.PreambleFunc != nil {
		r.Lexer.PreambleFunc(len(src), r.Options.Concurrency)
	}

	if r.Parser.PreambleFunc != nil {
		r.Parser.PreambleFunc(len(src), r.Options.Concurrency)
	}

	// Initialize Scanner and Grammar
	scanner := r.Lexer.Scanner(src, &r.Options)

	b.StartTimer()

	parser := r.Parser.Parser(src, &r.Options)

	b.StopTimer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokensLists, err := scanner.Lex(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not lex: %w", err)
	}

	b.StartTimer()

	token, err := parser.Parse(ctx, tokensLists)
	if err != nil {
		return nil, fmt.Errorf("could not parse: %w", err)
	}

	return token, nil
}

func Profile(t *testing.T,
	newLexer func() *gopapageno.Lexer, newGrammar func() *gopapageno.Grammar, opts *gopapageno.RunOptions, filename string) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("could not read source file %s: %v", filename, err)
	}

	r := gopapageno.NewRunner(
		newLexer(),
		newGrammar(),
		gopapageno.WithConcurrency(opts.Concurrency),
		gopapageno.WithAverageTokenLength(opts.AvgTokenLength),
		gopapageno.WithReductionStrategy(opts.ReductionStrategy),
		gopapageno.WithParallelFactor(opts.ParallelFactor),
		gopapageno.WithGarbageCollection(false),
	)

	ctx := context.Background()

	_, err = r.Run(ctx, bytes)
	if err != nil {
		t.Fatalf("could not parse source: %v", err)
	}
}
