package main

import (
	"flag"
	"fmt"
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
	lexerFlag := flag.String("l", "", "lexer source file")
	parserFlag := flag.String("g", "", "parser source file")
	outputFlag := flag.String("o", ".", "output directory")
	typesOnlyFlag := flag.Bool("types-only", false, "generate types only")

	logFlag := flag.Bool("log", false, "enable logging during generation")

	flag.Parse()

	if *lexerFlag == "" || *parserFlag == "" {
		return fmt.Errorf("lexer and parser files must be provided")
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
		Logger:                    log.New(logOut, "", 0),
	}

	if err := generator.Generate(opts); err != nil {
		return fmt.Errorf("could not generate: %w", err)
	}

	return nil
}