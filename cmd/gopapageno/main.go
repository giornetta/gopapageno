package main

import (
	"flag"
	"fmt"
	"github.com/giornetta/gopapageno"
	"io"
	"log"
	"os"

	"github.com/giornetta/gopapageno/generator"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	lexiconFlag := flag.String("l", "", "lexer source file")
	grammarFlag := flag.String("g", "", "grammar source file")
	outputFlag := flag.String("o", ".", "output directory")
	typesOnlyFlag := flag.Bool("types-only", false, "generate types only")
	benchmarkFlag := flag.Bool("benchmark", false, "generate benchmarks")

	strategyFlag := flag.String("s", "opp", "strategy to use during parser generation: opp/aopp/copp")

	logFlag := flag.Bool("log", false, "enable logging during generation")

	flag.Parse()

	if *lexiconFlag == "" || *grammarFlag == "" {
		return fmt.Errorf("lexicon and grammar description files must be provided")
	}

	strategy := gopapageno.OPP
	if *strategyFlag == "aopp" {
		strategy = gopapageno.AOPP
	} else if *strategyFlag == "copp" {
		strategy = gopapageno.COPP
	}

	var logOut io.Writer
	if *logFlag {
		logOut = os.Stderr
	} else {
		logOut = io.Discard
	}

	opts := &generator.Options{
		LexerDescriptionFilename:  *lexiconFlag,
		ParserDescriptionFilename: *grammarFlag,
		OutputDirectory:           *outputFlag,
		TypesOnly:                 *typesOnlyFlag,
		GenerateBenchmarks:        *benchmarkFlag,
		Strategy:                  strategy,
		Logger:                    log.New(logOut, "", 0),
	}

	if err := generator.Generate(opts); err != nil {
		return fmt.Errorf("could not generate: %w", err)
	}

	return nil
}
