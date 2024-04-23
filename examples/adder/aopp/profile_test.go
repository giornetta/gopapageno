package main

import (
	"context"
	"github.com/giornetta/gopapageno"
	"os"
	"path"
	"testing"
)

const baseFolder = "./data/"

const (
	fileSmall = "small.txt"
	fileMB    = "1MB.txt"
	file10MB  = "10MB.txt"
)

const (
	resultSmall = 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10
	resultMB    = (1 + 2 + 3 + 11 + 222 + 3333 + (1 + 2)) * 26000
	result10MB  = (1 + 2 + 3 + 11 + 222 + 3333 + (1 + 2)) * 260000
)

func TestProfile(t *testing.T) {
	c := 12
	strat := gopapageno.ReductionSweep
	file := path.Join(baseFolder, file10MB)

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

	if *root.Value.(*int64) != result10MB {
		t.Fatalf("wrong result: %v", *root.Value.(*int64))
	}
}
