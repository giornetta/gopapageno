package main

import (
	"context"
	"flag"
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
	"generated-1000.json": nil,
	"generated-2000.json": nil,
	"emojis-100.json":     nil,
}

var reductionFlag string

func TestMain(m *testing.M) {
	flag.StringVar(&reductionFlag, "s", "sweep", "parsing strategy to execute")

	flag.Parse()

	os.Exit(m.Run())
}

func BenchmarkParse(b *testing.B) {
	strat := gopapageno.ReductionSweep
	if reductionFlag == "parallel" {
		strat = gopapageno.ReductionParallel
	} else if reductionFlag == "mixed" {
		strat = gopapageno.ReductionMixed
	}

	threads := runtime.NumCPU()

	for filename, _ := range table {
		for c := 1; c <= threads; c = min(c*2, threads) {
			b.Run(fmt.Sprintf("%s/%dT", filename, c), func(b *testing.B) {
				r := gopapageno.NewRunner(
					NewLexer(),
					NewGrammar(),
					gopapageno.WithConcurrency(c),
					gopapageno.WithReductionStrategy(strat))

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
