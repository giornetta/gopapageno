package main

import (
	"fmt"
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
	"path"
	"runtime"
	"testing"
)

const baseFolder = "../data/"

const (
	fileSmall = "small.txt"
	fileMB    = "1MB.txt"
	file10MB  = "10MB.txt"
)

const (
	resultSmall = 1111 + 2222 + 33 + (44 + 555) + 6 + 7777
	resultMB    = (1 + 2 + 3 + 11 + 222 + 3333 + (1 + 2)) * 26000
	result10MB  = (1 + 2 + 3 + 11 + 222 + 3333 + (1 + 2)) * 260000
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
					gopapageno.WithPreallocFunc(ParserPreallocMem),
					gopapageno.WithReductionStrategy(gopapageno.ReductionParallel))

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
