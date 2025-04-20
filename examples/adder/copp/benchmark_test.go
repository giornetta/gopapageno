package main

import (
	"testing"

	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
)

const baseFolder = "../data/"

var entries = []*benchmark.Entry[int64]{
	{
		Filename:       baseFolder + "1MB.txt",
		ParallelFactor: 1,
		AvgTokenLength: 2,
		Result:         (1 + 2 + 3 + 11 + 222 + 3333 + (1 + 2)) * 26000,
	},
	{
		Filename:       baseFolder + "10MB.txt",
		ParallelFactor: 1,
		AvgTokenLength: 2,
		Result:         (1 + 2 + 3 + 11 + 222 + 3333 + (1 + 2)) * 260000,
	},
}

func BenchmarkParse(b *testing.B) {
	benchmark.Runner(b, gopapageno.COPP, gopapageno.ReductionParallel, NewLexer, NewGrammar, entries)
}

func BenchmarkParseOnly(b *testing.B) {
	benchmark.ParserRunner(b, gopapageno.COPP, gopapageno.ReductionParallel, NewLexer, NewGrammar, entries)
}
