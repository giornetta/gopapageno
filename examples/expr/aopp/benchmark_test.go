package main

import (
	"testing"

	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
)

const baseFolder = "../data/"

var entries = []*benchmark.Entry[int64]{
	{
		Filename:       baseFolder + "small.txt",
		ParallelFactor: 1,
		AvgTokenLength: 2,
		Result:         1 + 2*3*(4+5),
	},
	{
		Filename:       baseFolder + "1MB.txt",
		ParallelFactor: 1,
		AvgTokenLength: 2,
		Result:         (1*2*3 + 11*222*3333*(1+2)) * 25966,
	},
	{
		Filename:       baseFolder + "10MB.txt",
		ParallelFactor: 1,
		AvgTokenLength: 2,
		Result:         (1*2*3 + 11*222*3333*(1+2)) * 257473,
	},
}

func BenchmarkParse(b *testing.B) {
	benchmark.Runner(b, gopapageno.AOPP, gopapageno.ReductionParallel, NewLexer, NewGrammar, entries)
}

func BenchmarkParseOnly(b *testing.B) {
	benchmark.ParserRunner(b, gopapageno.AOPP, gopapageno.ReductionParallel, NewLexer, NewGrammar, entries)
}
