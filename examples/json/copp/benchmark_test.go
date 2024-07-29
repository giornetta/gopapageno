package main

import (
	"context"
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
	"os"
	"path"
	"runtime"
	"testing"
)

const baseFolder = "../data/"

var table = map[string]any{
	baseFolder + "emojis-555.json": nil,
}

func BenchmarkParse(b *testing.B) {
	benchmark.Runner[any](b, gopapageno.COPP, NewLexer, NewGrammar, table)
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
