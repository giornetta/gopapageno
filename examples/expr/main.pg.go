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
	start := time.Now()

	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(time.Since(start))
}

func run() error {
	sourceFlag := flag.String("f", "", "source file")
	concurrencyFlag := flag.Int("c", 1, "number of concurrent goroutines to spawn")
	logFlag := flag.Bool("log", false, "enable logging")

	flag.Parse()

	bytes, err := os.ReadFile(*sourceFlag)
	if err != nil {
		return fmt.Errorf("could not read source file %s: %w", *sourceFlag, err)
	}

	var logOut io.Writer
	if *logFlag {
		logOut = os.Stderr
	} else {
		logOut = io.Discard
	}

	p := NewParser(
		gopapageno.WithConcurrency(*concurrencyFlag),
		gopapageno.WithLogging(log.New(logOut, "", 0)))

	LexerPreallocMem(len(bytes), p.Concurrency())
	ParserPreallocMem(len(bytes), p.Concurrency())

	ctx := context.Background()

	root, err := p.Parse(ctx, bytes)
	if err != nil {
		return fmt.Errorf("could not parse source: %w", err)
	}

	// Added manually.
	fmt.Println(*root.Value.(*int64))

	return nil
}