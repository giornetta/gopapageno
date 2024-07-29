package benchmark

import (
	"context"
	"fmt"
	"github.com/giornetta/gopapageno"
	"os"
	"runtime"
	"testing"
)

func Runner[T any](b *testing.B, parsingStrategy gopapageno.ParsingStrategy, newLexer func() *gopapageno.Lexer, newGrammar func() *gopapageno.Grammar, table map[string]T) {
	reductionStrategies := []gopapageno.ReductionStrategy{gopapageno.ReductionSweep, gopapageno.ReductionParallel, gopapageno.ReductionMixed}

	threads := runtime.NumCPU()

	b.Run(fmt.Sprintf("strategy=%s", parsingStrategy), func(b *testing.B) {
		for filename, _ := range table {
			b.Run(fmt.Sprintf("file=%s", filename), func(b *testing.B) {
				for c := 1; c <= threads; c++ {
					b.Run(fmt.Sprintf("goroutines=%d", c), func(b *testing.B) {
						for _, reductionStrat := range reductionStrategies {
							b.Run(fmt.Sprintf("reduction=%s", reductionStrat), func(b *testing.B) {

								r := gopapageno.NewRunner(
									newLexer(),
									newGrammar(),
									gopapageno.WithConcurrency(c),
									gopapageno.WithReductionStrategy(reductionStrat))

								Run(b, r, filename)
							})
						}
					})
				}
			})
		}
	})

}

func RunExpect[T comparable](b *testing.B, r *gopapageno.Runner, filename string, expected T) {
	b.StopTimer()

	bytes, err := os.ReadFile(filename)
	if err != nil {
		b.Fatalf("could not read source file: %v", err)
	}

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

func Run(b *testing.B, r *gopapageno.Runner, filename string) {
	b.StopTimer()

	bytes, err := os.ReadFile(filename)
	if err != nil {
		b.Fatalf("could not read source file: %v", err)
	}

	ctx := context.Background()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := r.Run(ctx, bytes)
		if err != nil {
			b.Fatalf("could not parse source file: %v", err)
		}
	}
}
