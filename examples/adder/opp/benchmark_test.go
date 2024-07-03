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
	fileMB   = "1MB.txt"
	file10MB = "10MB.txt"
)

const (
	resultMB   = (1 + 2 + 3 + 11 + 222 + 3333 + (1 + 2)) * 26000
	result10MB = (1 + 2 + 3 + 11 + 222 + 3333 + (1 + 2)) * 260000
)

var table = map[string]int64{
	fileMB:   resultMB,
	file10MB: result10MB,
}

func BenchmarkParse(b *testing.B) {
	threads := runtime.NumCPU()

	for filename, expected := range table {
		for c := 1; c <= threads; c = min(c*2, threads) {
			b.Run(fmt.Sprintf("%s/%dT", filename, c), func(b *testing.B) {
				p := NewParser(
					gopapageno.WithConcurrency(c),
					gopapageno.WithPreallocFunc(ParserPreallocMem),
					gopapageno.WithReductionStrategy(gopapageno.ReductionMixed))

				b.ResetTimer()

				benchmark.RunExpect[int64](b, p, path.Join(baseFolder, filename), expected)
			})

			runtime.GC()

			if c == threads {
				break
			}
		}
	}
}

func TestProfile(t *testing.T) {
	c := 8
	strat := gopapageno.ReductionParallel

	filename := file10MB

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

	root, err := p.Parse(ctx, bytes)
	if err != nil {
		t.Fatalf("could not parse source: %v", err)
	}

	if *root.Value.(*int64) != table[filename] {
		t.Fatalf("wrong result: %v", *root.Value.(*int64))
	}
}
