package main

import (
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
	"runtime"
	"testing"
)

const baseFolder = "../data/"

var table = map[string]any{
	baseFolder + "generated-1000.json": nil,
	baseFolder + "emojis-100.json":     nil,
}

func BenchmarkParse(b *testing.B) {
	benchmark.Runner[any](b, gopapageno.AOPP, NewLexer, NewGrammar, table)
}

func TestProfile(t *testing.T) {
	c := runtime.NumCPU()
	avgLen := gopapageno.DefaultAverageTokenLength
	strat := gopapageno.ReductionParallel

	filename := ""

	benchmark.Profile(
		t,
		NewLexer, NewGrammar,
		c, avgLen, strat,
		filename)
}
