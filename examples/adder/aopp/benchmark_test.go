package main

import (
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
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
	benchmark.Runner[int64](b, gopapageno.AOPP, NewLexer, NewGrammar, table)
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
