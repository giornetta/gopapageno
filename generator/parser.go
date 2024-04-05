package generator

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
)

var (
	axiomRegexp = regexp.MustCompile("^%axiom\\s*([a-zA-Z][a-zA-Z0-9]*)\\s*$")
)

type rule struct {
	LHS    string
	RHS    []string
	Action string
}

func (r rule) String() string {
	return fmt.Sprintf("%s -> %s", r.LHS, strings.Join(r.RHS, " "))
}

type parserDescriptor struct {
	axiom    string
	preamble string
	rules    []rule

	// nonterminals is nil until inferTokens() is executed successfully.
	nonterminals *set[string]

	// terminals is nil until inferTokens() is executed successfully.
	terminals *set[string]

	precMatrix precedenceMatrix
}

func parseParserDescription(r io.Reader, logger *log.Logger) (*parserDescriptor, error) {
	logger.Printf("Parsing parser description file...\n")

	scanner := bufio.NewScanner(r)

	var preambleBuilder strings.Builder
	for scanner.Scan() {
		l := scanner.Bytes()
		if separatorRegexp.Match(l) {
			break
		}

		preambleBuilder.Write(l)
		preambleBuilder.WriteString("\n")
	}

	var axiom string
	moreThanOneAxiomWarning := false

	for scanner.Scan() {
		l := scanner.Text()
		if separatorRegexp.MatchString(l) {
			break
		}

		axiomMatch := axiomRegexp.FindStringSubmatch(l)
		if axiomMatch != nil {
			if axiom != "" && !moreThanOneAxiomWarning {
				fmt.Println("Warning: axiom is defined more than once")
				moreThanOneAxiomWarning = true
			}
			axiom = axiomMatch[1]
		}
	}

	if axiom == "" {
		return nil, fmt.Errorf("no axiom is defined")
	}

	logger.Printf("Axiom: %s\n", axiom)

	var sb strings.Builder
	for scanner.Scan() {
		l := scanner.Bytes()
		sb.Write(l)
		sb.WriteString("\n")
	}

	rules, err := parseRules(sb.String())
	if err != nil {
		return nil, fmt.Errorf("could not parse rules: %w", err)
	}

	logger.Printf("Parser Rules:\n")
	for _, rule := range rules {
		logger.Printf("%s\n", rule)
	}

	return &parserDescriptor{
		axiom:    axiom,
		preamble: preambleBuilder.String(),
		rules:    rules,
	}, nil
}

func parseRules(input string) ([]rule, error) {
	rules := make([]rule, 0)

	var pos int
	skipSpaces(input, &pos)

	for pos < len(input) {
		firstRule := rule{}

		lhs := getIdentifier(input, &pos)
		if lhs == "" {
			return nil, fmt.Errorf("missing or invalid identifier for lhs")
		}
		firstRule.LHS = lhs

		skipSpaces(input, &pos)

		if input[pos] == ':' {
			pos++
		} else {
			return nil, fmt.Errorf("rule %s is missing a colon between lhs and rhs", lhs)
		}

		skipSpaces(input, &pos)

		firstRule.RHS = make([]string, 0)

		for input[pos] != '{' {
			rhsToken := getIdentifier(input, &pos)
			if rhsToken == "" {
				return nil, fmt.Errorf("rule %s is missing an identifier for rhs", lhs)
			}
			firstRule.RHS = append(firstRule.RHS, rhsToken)

			skipSpaces(input, &pos)
		}

		semFun := getSemanticFunction(input, &pos)

		firstRule.Action = semFun

		rules = append(rules, firstRule)

		for {
			skipSpaces(input, &pos)

			if input[pos] == ';' {
				pos++
				break
			} else if input[pos] == '|' {
				pos++

				skipSpaces(input, &pos)

				nextRule := rule{}
				nextRule.LHS = lhs
				nextRule.RHS = make([]string, 0)

				for input[pos] != '{' {
					var rhsToken string
					rhsToken = getIdentifier(input, &pos)
					if rhsToken == "" {
						return nil, fmt.Errorf("rule %s is missing an identifier for rhs", lhs)
					}
					nextRule.RHS = append(nextRule.RHS, rhsToken)

					skipSpaces(input, &pos)
				}

				semFun := getSemanticFunction(input, &pos)

				nextRule.Action = semFun

				rules = append(rules, nextRule)
			} else {
				return nil, fmt.Errorf("invalid character at the end of rule %s", lhs)
			}
		}

		skipSpaces(input, &pos)
	}

	return rules, nil
}

// compile completes the parser description by doing all necessary checks and
// transformations in order to produce a correct OPG.
func (p *parserDescriptor) compile(logger *log.Logger) error {
	p.inferTokens()

	if !p.isAxiomUsed() {
		return fmt.Errorf("axiom isn't used in any rule")
	}

	p.deleteRepeatedRHS()

	precMatrix, err := p.newPrecedenceMatrix()
	if err != nil {
		return fmt.Errorf("could not create precedence matrix: %w", err)
	}
	p.precMatrix = precMatrix

	p.sortRulesByRHS()

	return nil
}

// inferTokens populates the two sets nonterminals and terminals
// with tokens found in the grammar rules.
func (p *parserDescriptor) inferTokens() {
	p.nonterminals = newSet[string]()
	tokens := newSet[string]()

	for _, rule := range p.rules {
		if !p.nonterminals.Contains(rule.LHS) {
			p.nonterminals.Add(rule.LHS)
		}
		for _, token := range rule.RHS {
			if !tokens.Contains(token) {
				tokens.Add(token)
			}
		}
	}

	p.terminals = tokens.Difference(p.nonterminals)

	// TODO: Change this?
	p.nonterminals.Add("_EMPTY")
	p.terminals.Add("_TERM")
}

// isAxiomUsed checks if the axiom is present in any rules' LHS.
func (p *parserDescriptor) isAxiomUsed() bool {
	for _, rule := range p.rules {
		if rule.LHS == p.axiom {
			return true
		}
	}

	return false
}

func (p *parserDescriptor) emit(f io.Writer, opts *Options) {
	/************
	 * Preamble *
	 ************/
	fmt.Fprintf(f, p.preamble)
	fmt.Fprintf(f, "\n\n")

	p.emitTokens(f)

	fmt.Fprintf(f, "\nfunc NewParser(opts ...gopapageno.ParserOpt) *gopapageno.Parser {\n")

	/*****************
	 * Token Numbers *
	 *****************/
	fmt.Fprintf(f, "\tnumTerminals := uint16(%d)\n", p.terminals.Len())
	fmt.Fprintf(f, "\tnumNonTerminals := uint16(%d)\n\n", p.nonterminals.Len())

	/*********
	 * Rules *
	 *********/
	maxRHSLen := 0
	for _, rule := range p.rules {
		ruleLen := len(rule.RHS)
		if ruleLen > maxRHSLen {
			maxRHSLen = ruleLen
		}
	}

	fmt.Fprintf(f, "\tmaxRHSLen := %d\n", maxRHSLen)
	fmt.Fprint(f, "\trules := []gopapageno.Rule{\n")
	for _, rule := range p.rules {
		fmt.Fprintf(f, "\t\t{%s, []gopapageno.TokenType{%s}},\n", rule.LHS, strings.Join(rule.RHS, ", "))
	}
	fmt.Fprintf(f, "\t}\n")

	trie, err := newTrie(p.rules, p.nonterminals, p.terminals)
	if err != nil {
		// TODO: Change this handling.
		panic(err)
	}

	compressedTrie := trie.Compress(p.nonterminals, p.terminals)

	fmt.Fprintf(f, "\tcompressedRules := []uint16{")
	if len(compressedTrie) > 0 {
		fmt.Fprintf(f, "%d", compressedTrie[0])
		for i := 1; i < len(compressedTrie); i++ {
			fmt.Fprintf(f, ", %d", compressedTrie[i])
		}
	}
	fmt.Fprintf(f, "\t}\n\n")

	/*********************
	 * Precedence Matrix *
	 *********************/
	fmt.Fprintf(f, "\tprecMatrix := [][]gopapageno.Precedence{\n")
	for i, _ := range p.precMatrix {
		fmt.Fprintf(f, "\t\t{")

		for j, _ := range p.precMatrix {
			if j < len(p.precMatrix)-1 {
				fmt.Fprintf(f, "gopapageno.Prec%s, ", p.precMatrix[i][j].String())
			} else {
				fmt.Fprintf(f, "gopapageno.Prec%s", p.precMatrix[i][j].String())
			}
		}
		fmt.Fprintf(f, "},\n")
	}
	fmt.Fprintf(f, "\t}\n")

	bitPackedMatrix := bitPack(p.precMatrix)
	fmt.Fprintf(f, "\tbitPackedMatrix := []uint64{\n\t\t")
	for _, v := range bitPackedMatrix {
		fmt.Fprintf(f, "%d, ", v)
	}
	fmt.Fprintf(f, "\n\t}\n\n")

	/*******************
	 * Parser Function *
	 *******************/
	fmt.Fprintf(f, "\tfn := func(rule uint16, lhs *gopapageno.Token, rhs []*gopapageno.Token, thread int){\n")
	fmt.Fprintf(f, "\t\tswitch rule {\n")
	for i, rule := range p.rules {
		fmt.Fprintf(f, "\t\tcase %d:\n", i)
		fmt.Fprintf(f, "\t\t\t%s0 := lhs\n", rule.LHS)
		for j, _ := range rule.RHS {
			fmt.Fprintf(f, "\t\t\t%s%d := rhs[%d]\n", rule.RHS[j], j+1, j)
		}
		fmt.Fprintf(f, "\n")

		if len(rule.RHS) > 0 {
			fmt.Fprintf(f, "\t\t\t%s0.Child = %s1\n", rule.LHS, rule.RHS[0])
			for j := 0; j < len(rule.RHS)-1; j++ {
				fmt.Fprintf(f, "\t\t\t%s%d.Next = %s%d\n", rule.RHS[j], j+1, rule.RHS[j+1], j+2)
			}
		}
		fmt.Fprintf(f, "\n")

		action := rule.Action
		action = strings.Replace(action, "$$", rule.LHS+"0", -1)
		for j, _ := range rule.RHS {
			action = strings.Replace(action, fmt.Sprintf("$%d", j+1), fmt.Sprintf("%s%d", rule.RHS[j], j+1), -1)
		}
		lines := strings.Split(action, "\n")
		for _, line := range lines {
			fmt.Fprintf(f, "\t\t\t")
			fmt.Fprintf(f, line)
			fmt.Fprintf(f, "\n")
		}
	}
	fmt.Fprintf(f, "\t\t}\n")
	fmt.Fprintf(f, "\t}\n\n")

	/********************
	 * Construct Parser *
	 ********************/
	fmt.Fprintf(f, "\treturn gopapageno.NewParser(\n")
	fmt.Fprintf(f, "\t\tNewLexer(),\n")
	fmt.Fprintf(f, "\t\tnumTerminals,\n\t\tnumNonTerminals,\n")
	fmt.Fprintf(f, "\t\tmaxRHSLen,\n")
	fmt.Fprintf(f, "\t\trules,\n")
	fmt.Fprintf(f, "\t\tcompressedRules,\n")
	fmt.Fprintf(f, "\t\tprecMatrix,\n")
	fmt.Fprintf(f, "\t\tbitPackedMatrix,\n")
	fmt.Fprintf(f, "\t\tfn,\n")
	fmt.Fprintf(f, "\t\topts...)\n}\n\n")
}

func (p *parserDescriptor) emitTokens(f io.Writer) {
	fmt.Fprintf(f, "// Non-terminals\n")
	fmt.Fprintf(f, "const (\n")
	for i, token := range p.nonterminals.Slice() {
		if token == "_EMPTY" {
			continue
		}

		if i == 0 {
			fmt.Fprintf(f, "\t%s = gopapageno.TokenEmpty + 1 + iota\n", token)
		} else {
			fmt.Fprintf(f, "\t%s\n", token)
		}
	}
	fmt.Fprintf(f, ")\n\n")

	fmt.Fprintf(f, "// Terminals\n")
	fmt.Fprintf(f, "const (\n")
	for i, token := range p.terminals.Slice() {
		if token == "_TERM" {
			continue
		}

		if i == 0 {
			fmt.Fprintf(f, "\t%s = gopapageno.TokenTerm + 1 + iota\n", token)
		} else {
			fmt.Fprintf(f, "\t%s\n", token)
		}
	}
	fmt.Fprintf(f, ")\n\n")
}
