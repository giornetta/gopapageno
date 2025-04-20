package main

import (
	"testing"

	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
)

const baseFolder = "../data/"

var entries = []*benchmark.Entry[any]{
	{
		Filename:       baseFolder + "citylots.json",
		ParallelFactor: 0.5,
		AvgTokenLength: 4,
		Result:         nil,
	},
	{
		Filename:       baseFolder + "emojis-1000.json",
		ParallelFactor: 1,
		AvgTokenLength: 8,
		Result:         nil,
	},
	{
		Filename:       baseFolder + "wikidata-lexemes.json",
		ParallelFactor: 0,
		AvgTokenLength: 4,
		Result:         nil,
	},
}

func BenchmarkParse(b *testing.B) {
	benchmark.Runner(b, gopapageno.AOPP, gopapageno.ReductionParallel, NewLexer, NewGrammar, entries)
}

func BenchmarkParseOnly(b *testing.B) {
	benchmark.ParserRunner(b, gopapageno.AOPP, gopapageno.ReductionParallel, NewLexer, NewGrammar, entries)
}
