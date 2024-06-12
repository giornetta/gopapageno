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

const (
	file1000 = "generated-1000.json"
	file2000 = "generated-2000.json"
)

var table = []string{
	file1000,
	file2000,
}

func BenchmarkParse(b *testing.B) {
	threads := runtime.NumCPU()

	for _, filename := range table {
		for c := 1; c <= threads; c = min(c*2, threads) {
			b.Run(fmt.Sprintf("%s/%dT", filename, c), func(b *testing.B) {
				p := NewParser(
					gopapageno.WithConcurrency(c),
					gopapageno.WithPreallocFunc(ParserPreallocMem),
					gopapageno.WithReductionStrategy(gopapageno.ReductionSweep))

				b.ResetTimer()

				benchmark.Run(b, p, path.Join(baseFolder, filename))
			})

			runtime.GC()

			if c == threads {
				break
			}
		}
	}
}

func TestProfile(t *testing.T) {
	c := 12
	strat := gopapageno.ReductionParallel

	filename := "generated-2000.json"

	file := path.Join(baseFolder, filename)

	bytes, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("could not read source file %s: %v", file, err)
	}

	p := NewParser(
		gopapageno.WithConcurrency(c),
		gopapageno.WithPreallocFunc(ParserPreallocMem),
		gopapageno.WithReductionStrategy(strat),
	)

	ctx := context.Background()

	_, err = p.Parse(ctx, bytes)
	if err != nil {
		t.Fatalf("could not parse source: %v", err)
	}
}
