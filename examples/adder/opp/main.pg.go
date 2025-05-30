// Code generated by Gopapageno.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/giornetta/gopapageno"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	start := time.Now()

	sourceFlag := flag.String("f", "", "source file")
	concurrencyFlag := flag.Int("c", 1, "number of concurrent goroutines to spawn")
	strategyFlag := flag.String("s", "sweep", "parsing strategy to execute")
	logFlag := flag.Bool("log", false, "enable logging")
	avgTokensFlag := flag.Int("avg", 4, "average length of tokens")

	cpuProfileFlag := flag.String("cpuprof", "", "output file for CPU profiling")
	memProfileFlag := flag.String("memprof", "", "output file for Memory profiling")

	flag.Parse()

	bytes, err := os.ReadFile(*sourceFlag)
	if err != nil {
		return fmt.Errorf("could not read source file %s: %w", *sourceFlag, err)
	}

	logOut := io.Discard
	if *logFlag {
		logOut = os.Stderr
	}

	cpuProfileWriter := io.Discard
	if *cpuProfileFlag != "" {
		cpuProfileWriter, err = os.Create(*cpuProfileFlag)
		if err != nil {
			cpuProfileWriter = io.Discard
		}
	}

	memProfileWriter := io.Discard
	if *memProfileFlag != "" {
		memProfileWriter, err = os.Create(*memProfileFlag)
		if err != nil {
			memProfileWriter = io.Discard
		}
	}

	strat := gopapageno.ReductionSweep
	if *strategyFlag == "parallel" {
		strat = gopapageno.ReductionParallel
	} else if *strategyFlag == "mixed" {
		strat = gopapageno.ReductionMixed
	}

	r := gopapageno.NewRunner(
		NewLexer(),
		NewGrammar(),
		gopapageno.WithConcurrency(*concurrencyFlag),
		gopapageno.WithLogging(log.New(logOut, "", 0)),
		gopapageno.WithCPUProfiling(cpuProfileWriter),
		gopapageno.WithMemoryProfiling(memProfileWriter),
		gopapageno.WithReductionStrategy(strat),
		gopapageno.WithAverageTokenLength(*avgTokensFlag),
	)

	ctx := context.Background()

	root, err := r.Run(ctx, bytes)
	if err != nil {
		return fmt.Errorf("could not parse source: %w", err)
	}

	fmt.Printf("Parsing took: %v\n", time.Since(start))
	fmt.Printf("Result: %v\n", *root.Value.(*int64))

	h := root.Height()
	s := root.Size()
	fmt.Printf("Height: %d\nSize: %d\n", h, s)
	if h < 10 && s < 100 {
		fmt.Println(SprintToken[any](root))
	}

	return nil
}
