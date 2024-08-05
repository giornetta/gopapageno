package main

import (
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
	"path"
	"runtime"
	"testing"
)

const baseFolder = "../data/"

var table = map[string]any{}

func BenchmarkParse(b *testing.B) {
	benchmark.Runner[any](b, gopapageno.COPP, NewLexer, NewGrammar, table)
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
