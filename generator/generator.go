package generator

import (
	"fmt"
	"os"

	"github.com/giornetta/gopapageno/generator/regex"
)

func Generate(lexerFilename string, parserFilename string, outdir string) error {
	lexerFile, err := os.Open(lexerFilename)
	if err != nil {
		return fmt.Errorf("could not open lexer file: %v", err)
	}
	defer lexerFile.Close()

	lexRules, cutPoints, lexCode := parseLexer(lexerFile)

	fmt.Printf("Lex rules (%d):\n", len(lexRules))
	for _, r := range lexRules {
		fmt.Println(r)
	}

	fmt.Printf("Cut points regex: %s\n", cutPoints)

	fmt.Println("Lex code:")
	fmt.Println(lexCode)

	var dfa regex.Dfa

	if len(lexRules) == 0 {
		return fmt.Errorf("lexer doesn't contain any rules")
	}

	var nfa *regex.Nfa
	success, result := regex.ParseString([]byte(lexRules[0].Regex), 1)
	if success {
		nfa = result.Value.(*regex.Nfa)
		nfa.AddAssociatedRule(0)
	} else {
		return fmt.Errorf("could not parse regular expression %v", lexRules[0].Regex)
	}
	for i := 1; i < len(lexRules); i++ {
		var curNfa *regex.Nfa
		success, result = regex.ParseString([]byte(lexRules[i].Regex), 1)
		if success {
			curNfa = result.Value.(*regex.Nfa)
			curNfa.AddAssociatedRule(i)
			nfa.Unite(*curNfa)
		} else {
			return fmt.Errorf("could not parse regular expression %v", lexRules[0].Regex)
		}
	}

	dfa = nfa.ToDfa()

	/*ok, hasRuleNum, ruleNum := dfa.Check([]byte(" "))
	if ok {
		fmt.Println("Ok")
		if hasRuleNum {
			fmt.Println("RuleNum:", ruleNum)
		} else {
			fmt.Println("No rule")
		}
	} else {
		fmt.Println("Not ok")
	}*/

	var cutPointsDfa regex.Dfa
	if cutPoints == "" {
		cutPointsNfa := regex.NewEmptyStringNfa()
		cutPointsDfa = cutPointsNfa.ToDfa()
	} else {
		var cutPointsNfa *regex.Nfa
		success, result := regex.ParseString([]byte(cutPoints), 1)
		if success {
			cutPointsNfa = result.Value.(*regex.Nfa)
		} else {
			return fmt.Errorf("could not parse regular expression %v", lexRules[0].Regex)
		}
		cutPointsDfa = cutPointsNfa.ToDfa()
	}

	parserPreamble, axiom, rules := parseGrammar(parserFilename)

	fmt.Println("Go preamble:")
	fmt.Println(parserPreamble)

	if axiom == "" {
		return fmt.Errorf("axiom is not defined")
	} else {
		fmt.Println("Axiom:", axiom)
	}

	fmt.Printf("Rules (%d):\n", len(rules))
	for _, r := range rules {
		fmt.Println(r)
	}

	nonterminals, terminals := inferTokens(rules)

	fmt.Printf("Nonterminals (%d): %s\n", len(nonterminals), nonterminals)
	fmt.Printf("Terminals (%d): %s\n", len(terminals), terminals)

	if !checkAxiomUsage(rules, axiom) {
		return fmt.Errorf("axiom is not used in any rule")
	}

	newRules, newNonterminals := deleteRepeatedRHS(nonterminals, terminals, axiom, rules)

	fmt.Printf("New rules after elimination of repeated rhs (%d):\n", len(newRules))
	for _, r := range newRules {
		fmt.Println(r)
	}

	fmt.Printf("New nonterminals (%d): %s\n", len(newNonterminals), newNonterminals)

	precMatrix, err := createPrecMatrix(newRules, newNonterminals, terminals)
	if err != nil {
		return err
	}

	fmt.Println("Precedence matrix:")
	fmt.Println(precMatrix)

	sortedRules := sortRulesByRHS(newRules, newNonterminals, terminals)
	fmt.Printf("Sorted rules (%d):\n", len(sortedRules))
	for _, r := range sortedRules {
		fmt.Println(r)
	}

	err = emitOutputFolder(outdir)
	handleEmissionError(err)
	err = emitLexerFunction(outdir, lexCode, lexRules)
	handleEmissionError(err)
	err = emitLexerAutomata(outdir, dfa, cutPointsDfa)
	handleEmissionError(err)
	err = emitTokens(outdir, newNonterminals, terminals)
	handleEmissionError(err)
	err = emitRules(outdir, sortedRules, newNonterminals, terminals)
	handleEmissionError(err)
	err = emitFunction(outdir, parserPreamble, sortedRules)
	handleEmissionError(err)
	err = emitPrecMatrix(outdir, terminals, precMatrix)
	handleEmissionError(err)
	err = emitCommonFiles(outdir)
	handleEmissionError(err)

	return nil
}

func handleEmissionError(e error) {
	if e != nil {
		fmt.Println(e.Error())
	}
}
