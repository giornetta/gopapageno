package main

import (
	"flag"
	"fmt"
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
	var lexerFlag = flag.String("l", "", "lexer source file")
	var parserFlag = flag.String("g", "", "parser source file")
	var outputFlag = flag.String("o", ".", "output directory")

	flag.Parse()

	if *lexerFlag == "" || *parserFlag == "" {
		return fmt.Errorf("lexer and parser files must be provided")
	}

	generator.Generate(*lexerFlag, *parserFlag, *outputFlag)
	return nil
}
