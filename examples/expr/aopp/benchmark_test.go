package main

import (
	"fmt"
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
	"path"
	"runtime"
	"testing"
)

const baseFolder = "./data/"

const (
	fileSmall    = "small.txt"
	fileMB       = "1MB.txt"
	file10MB     = "10MB.txt"
	fleLeftAssoc = "LeftAssoc.txt"
)

const (
	resultSmall = 1 + 2*3*(4+5)
	resultMB    = (1*2*3 + 11*222*3333*(1+2)) * 25966
	result10MB  = (1*2*3 + 11*222*3333*(1+2)) * 257473

	resultLeftAssoc = (1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9) * 50000
)

var table = []struct {
	filename string
	expected int64
}{
	{fileSmall, resultSmall},
	{fileMB, resultMB},
	{file10MB, result10MB},
}

func BenchmarkParse(b *testing.B) {
	threads := runtime.NumCPU()

	for _, v := range table {
		for c := 1; c <= threads; c = min(c*2, threads) {
			b.Run(fmt.Sprintf("%s/%dT", v.filename, c), func(b *testing.B) {
				p := NewParser(
					gopapageno.WithConcurrency(c),
					gopapageno.WithPreallocFunc(ParserPreallocMem))

				b.ResetTimer()

				benchmark.Run[int64](b, p, path.Join(baseFolder, v.filename), v.expected)
			})

			runtime.GC()

			if c == threads {
				break
			}
		}
	}
}

func BenchmarkLeftAssoc(b *testing.B) {
	threads := runtime.NumCPU()

	for c := 1; c <= threads; c = min(c*2, threads) {
		b.Run(fmt.Sprintf("%dT", c), func(b *testing.B) {
			p := NewParser(
				gopapageno.WithConcurrency(c),
				gopapageno.WithPreallocFunc(ParserPreallocMem))

			b.ResetTimer()

			benchmark.Run[int64](b, p, path.Join(baseFolder, fleLeftAssoc), resultLeftAssoc)
		})

		runtime.GC()

		if c == threads {
			break
		}
	}
}
