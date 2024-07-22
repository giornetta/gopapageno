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

var table = map[string]any{}

func BenchmarkParse(b *testing.B) {
	threads := runtime.NumCPU()

	for filename, _ := range table {
		for c := 1; c <= threads; c = min(c*2, threads) {
			b.Run(fmt.Sprintf("%s/%dT", filename, c), func(b *testing.B) {
				r := gopapageno.NewRunner(
					NewLexer(),
					NewGrammar(),
					gopapageno.WithConcurrency(c),
					gopapageno.WithReductionStrategy(gopapageno.ReductionSweep))

				b.ResetTimer()

				benchmark.Run(b, r, path.Join(baseFolder, filename))
			})

			runtime.GC()

			if c == threads {
				break
			}
		}
	}
}

func TestProfile(t *testing.T) {
	c := runtime.NumCPU()
	avgLen := gopapageno.DefaultAverageTokenLength
	strat := gopapageno.ReductionParallel

	var filename string

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
