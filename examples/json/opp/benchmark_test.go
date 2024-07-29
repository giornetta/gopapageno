package main

import (
	"context"
	"fmt"
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
	"os"
	"path"
	"runtime"
	"testing"
)

const baseFolder = "../data/"

var table = map[string]any{
	"emojis-100.json": nil,
}

func BenchmarkParse(b *testing.B) {
	reductionStrategies := []gopapageno.ReductionStrategy{gopapageno.ReductionSweep, gopapageno.ReductionParallel, gopapageno.ReductionMixed}

	threads := runtime.NumCPU()

	b.Run("strategy=opp", func(b *testing.B) {
		for filename, _ := range table {
			b.Run(fmt.Sprintf("file=%s", filename), func(b *testing.B) {
				for c := 1; c <= threads; c++ {
					b.Run(fmt.Sprintf("goroutines=%d", c), func(b *testing.B) {
						for _, strat := range reductionStrategies {
							b.Run(fmt.Sprintf("reduction=%s", strat), func(b *testing.B) {
								r := gopapageno.NewRunner(
									NewLexer(),
									NewGrammar(),
									gopapageno.WithConcurrency(c),
									gopapageno.WithReductionStrategy(strat))

								benchmark.Run(b, r, path.Join(baseFolder, filename))
							})
						}
					})
				}
			})
		}
	})
}

func TestProfile(t *testing.T) {
	c := runtime.NumCPU()
	avgLen := gopapageno.DefaultAverageTokenLength
	strat := gopapageno.ReductionParallel

	var filename string = "small.json"

	file := path.Join(baseFolder, filename)

	bytes, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("could not read source file %s: %v", file, err)
	}

	r := gopapageno.NewRunner(
		NewLexer(),
		NewGrammar(),
		gopapageno.WithConcurrency(c),
		gopapageno.WithAverageTokenLength(avgLen),
		gopapageno.WithReductionStrategy(strat),
	)

	ctx := context.Background()

	_, err = r.Run(ctx, bytes)
	if err != nil {
		t.Fatalf("could not parse source: %v", err)
	}
}
