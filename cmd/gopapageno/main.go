package main

import (
	"flag"
	"fmt"
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/generator"
	"io"
	"log"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()

		os.Exit(1)
	}
}

func run() error {
	lexerFlag := flag.String("l", "", "Lexer specification source file")
	parserFlag := flag.String("p", "", "Parser specification source file")

	outputFlag := flag.String("out", ".", "Output directory for generated files")
	typesOnlyFlag := flag.Bool("no-main", false, "Do not generate a main program for the parser")
	benchmarkFlag := flag.Bool("benchmark", false, "Generate a benchmark and profiling file")

	strategyFlag := flag.String("strat", "opp", "Strategy to use during parser generation: opp/aopp/copp")

	logFlag := flag.Bool("v", false, "Enables verbose logging during parser generation")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "GoPAPAGENO: generate parallel parsers based on Floyd's Operator Precedence Grammars.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: gopapageno [flags]\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *lexerFlag == "" || *parserFlag == "" {
		return fmt.Errorf("lexer and parser files must be provided")
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
		LexerDescriptionFilename:  *lexerFlag,
		ParserDescriptionFilename: *parserFlag,
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
