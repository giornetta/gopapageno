package generator

import (
	"bufio"
	"fmt"
	"github.com/giornetta/gopapageno"
	"io"
	"log"
	"regexp"
	"strings"
)

var (
	axiomRegexp = regexp.MustCompile("^%axiom\\s*([a-zA-Z][a-zA-Z0-9]*)\\s*$")
)

type parserDescriptor struct {
	axiom    string
	rules    []rule
	prefixes [][]string
	code     string

	// nonterminals is nil until inferTokens() is executed successfully.
	nonterminals *set[string]

	// terminals is nil until inferTokens() is executed successfully.
	terminals *set[string]

	precMatrix precedenceMatrix
}

func parseParserDescription(r io.Reader, opts *Options) (*parserDescriptor, error) {
	opts.Logger.Printf("Parsing parser description file...\n")

	scanner := bufio.NewScanner(r)

	// Parse axiom definition
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
				log.Println("Warning: axiom is defined more than once.")
				moreThanOneAxiomWarning = true
			}
			axiom = axiomMatch[1]
		}
	}

	if axiom == "" {
		return nil, fmt.Errorf("no axiom is defined")
	}

	opts.Logger.Printf("Axiom: %s\n", axiom)

	// Parse rules
	var sb strings.Builder
	for scanner.Scan() {
		l := scanner.Bytes()
		if separatorRegexp.Match(l) {
			break
		}

		sb.Write(l)
		sb.WriteString("\n")
	}

	rules, prefixes, err := parseRules(sb.String(), opts.Strategy)
	if err != nil {
		return nil, fmt.Errorf("could not parse rules: %w", err)
	}

	opts.Logger.Printf("Parser Rules:\n")
	for _, rule := range rules {
		opts.Logger.Printf("%s\n", rule)
	}

	// Parse code
	var preambleBuilder strings.Builder
	for scanner.Scan() {
		l := scanner.Bytes()

		preambleBuilder.Write(l)
		preambleBuilder.WriteString("\n")
	}

	return &parserDescriptor{
		axiom:    axiom,
		code:     preambleBuilder.String(),
		rules:    rules,
		prefixes: prefixes,
	}, nil
}

func parseRules(input string, strategy gopapageno.ParsingStrategy) ([]rule, [][]string, error) {
	rules := make([]rule, 0)
	prefixes := make([][]string, 0)

	var pos int
	skipSpaces(input, &pos)

	var lhs string
	for pos < len(input) {
		var rule rule
		rule.Type = gopapageno.RuleSimple

		// If we're reading an "alternate rule"
		if lhs == "" {
			// Read LHS
			lhs = getIdentifier(input, &pos)
			if lhs == "" {
				return nil, nil, fmt.Errorf("missing or invalid identifier for lhs")
			}
			skipSpaces(input, &pos)

			// Read production delimiter
			if input[pos] == ':' {
				pos++
			} else {
				return nil, nil, fmt.Errorf("rule %s is missing a colon between lhs and rhs", lhs)
			}
			skipSpaces(input, &pos)
		}
		rule.LHS = lhs

		// Read Rhs
		rule.RHS = make([]string, 0)

		rightSides := make([][]string, 1)

		for input[pos] != '{' {
			var rhsToken string
			if strategy != gopapageno.COPP {
				rhsToken = getIdentifier(input, &pos)
				if rhsToken == "" {
					return nil, nil, fmt.Errorf("rule %s is missing an identifier for rhs", lhs)
				}

				rule.RHS = append(rule.RHS, rhsToken)
			} else {
				if input[pos] == '(' {
					// If the next section is a ()+ part, get the list of all produced alternatives (even nested).
					_, alternatives, err := getAlternatives(input, &pos)
					if err != nil {
						return nil, nil, fmt.Errorf("rule %s is missing an alternative body for lhs", lhs)
					}

					//// Add each produced alternative to every rhs found so far.
					newRightSides := make([][]string, len(rightSides)*len(alternatives))
					for i := 0; i < len(rightSides); i++ {
						//newRightSides[i*len(alternatives)] = append(rightSides[i], flattened...)
						//
						//newRightSides[i*len(alternatives)+1] = append(rightSides[i], lhs)
						//newRightSides[i*len(alternatives)+1] = append(newRightSides[i*len(alternatives)+1], flattened[1:]...)
						//
						//newRightSides[i*len(alternatives)+2] = append(rightSides[i], lhs)
						//newRightSides[i*len(alternatives)+2] = append(newRightSides[i*len(alternatives)+2], flattened[1:]...)

						for j := 0; j < len(alternatives); j++ {
							newRightSides[i*len(alternatives)+j] = append(rightSides[i], alternatives[j]...)
						}
					}
					rightSides = newRightSides

					//prefixes = append(prefixes, alternatives...)

					//rule.RHS = append(rule.RHS, flattened...)
					rule.Type = gopapageno.RuleCyclic
				} else {
					// Get a simple identifier
					rhsToken = getIdentifier(input, &pos)
					if rhsToken == "" {
						return nil, nil, fmt.Errorf("rule %s is missing an identifier for rhs", lhs)
					}

					if len(rightSides) == 1 {
						rightSides[0] = append(rightSides[0], rhsToken)
					} else {
						//newRightSides := make([][]string, len(rightSides)*2)
						for i := 0; i < len(rightSides); i++ {
							rightSides[i] = append(rightSides[i], rhsToken)
							//rightSides[i+len(rightSides)] = append(rightSides[i], rhsToken)
							//
						}
						//rightSides[len(rightSides)-1] = append(rightSides[len(rightSides)-1], lhs)
						//rightSides = newRightSides
					}
					//rule.RHS = append(rule.RHS, rhsToken)
				}
			}

			skipSpaces(input, &pos)
		}

		semFun := getSemanticFunction(input, &pos)
		rule.Action = semFun

		if strategy != gopapageno.COPP {
			rule.Type = gopapageno.RuleSimple
			rules = append(rules, rule)
		} else {
			// rules = append(rules, rule)

			for i, rhs := range rightSides {
				rule.RHS = rhs

				if i == 0 {
					if len(rightSides) == 1 {
						rule.Type = gopapageno.RuleSimple
					} else {
						rule.Type = gopapageno.RuleCyclic
					}
				} else {
					rule.Type = gopapageno.RulePrefix
				}

				rules = append(rules, rule)
			}
		}
		//rules = append(rules, rule)

		skipSpaces(input, &pos)

		// rule.Type = gopapageno.RuleSimple
		if input[pos] == ';' {
			// We're done with rules with this lhs
			// Reset current lhs
			lhs = ""
		} else if input[pos] == '|' {
			// We have another rule with the same lhs
		} else {
			return nil, nil, fmt.Errorf("invalid character at the end of rule %s", lhs)
		}

		pos++
		skipSpaces(input, &pos)
	}

	return rules, prefixes, nil
}

// compile completes the parser description by doing all necessary checks and
// transformations in order to produce a correct OPG.
func (p *parserDescriptor) compile(opts *Options) error {
	p.inferTokens()

	if !p.isAxiomUsed() {
		return fmt.Errorf("axiom isn't used in any rule")
	}

	p.deleteRepeatedRHS()

	var precMatrix precedenceMatrix
	var err error

	precMatrix, err = p.newPrecedenceMatrix(opts)
	if err != nil {
		return fmt.Errorf("could not create precedence matrix: %w", err)
	}
	p.precMatrix = precMatrix

	if opts.Strategy == gopapageno.COPP {
		p.makeCyclicRules()
	}

	p.sortRulesByRHS()

	return nil
}

func (p *parserDescriptor) makeCyclicRules() {
	rules := make([]rule, 0)

	for _, r := range p.rules {
		if r.Type == gopapageno.RuleSimple {
			rules = append(rules, r)
			continue
		}

		if r.Type == gopapageno.RulePrefix {
			continue
		}

		// E -> T + T
		rules = append(rules, r)

		// E -> T + E
		rhs := make([]string, len(r.RHS))
		for i := 0; i < len(rhs)-1; i++ {
			rhs[i] = r.RHS[i]
		}
		rhs[len(rhs)-1] = r.LHS

		rules = append(rules, rule{
			LHS:    r.LHS,
			RHS:    rhs,
			Action: r.Action,
			Type:   gopapageno.RuleAppendLeft,
		})

		// E -> E + T
		rhs = make([]string, len(r.RHS))
		rhs[0] = r.LHS
		for i := 1; i < len(rhs); i++ {
			rhs[i] = r.RHS[i]
		}

		rules = append(rules, rule{
			LHS:    r.LHS,
			RHS:    rhs,
			Action: r.Action,
			Type:   gopapageno.RuleAppendRight,
		})

		// E -> E + E
		rhs = make([]string, len(r.RHS))
		rhs[0] = r.LHS
		for i := 1; i < len(rhs)-1; i++ {
			rhs[i] = r.RHS[i]
		}
		rhs[len(rhs)-1] = r.LHS

		rules = append(rules, rule{
			LHS:    r.LHS,
			RHS:    rhs,
			Action: r.Action,
			Type:   gopapageno.RuleCombine,
		})
	}

	p.rules = rules
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
	fmt.Fprintf(f, p.code)
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
		if len(rule.RHS) > maxRHSLen {
			maxRHSLen = len(rule.RHS)
		}
	}

	fmt.Fprintf(f, "\tmaxRHSLen := %d\n", maxRHSLen)
	fmt.Fprint(f, "\trules := []gopapageno.Rule{\n")
	for _, rule := range p.rules {
		fmt.Fprintf(f, "\t\t{%s, []gopapageno.TokenType{%s}, gopapageno.%s},\n", rule.LHS, strings.Join(rule.RHS, ", "), rule.Type)
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

	///*****************
	// * COPP Prefixes *
	// *****************/
	//maxPrefixLen := 0
	//for _, prefix := range p.prefixes {
	//	if len(prefix) > maxPrefixLen {
	//		maxPrefixLen = len(prefix)
	//	}
	//}
	//
	//fmt.Fprintf(f, "\tmaxPrefixLen := %d\n", maxPrefixLen)
	//fmt.Fprint(f, "\tprefixes := [][]gopapageno.TokenType{\n")
	//for _, prefix := range p.prefixes {
	//	fmt.Fprintf(f, "\t\t{%s},\n", strings.Join(prefix, ", "))
	//}
	//fmt.Fprintf(f, "\t}\n")

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
	p.emitParserFunctions(f)

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
	fmt.Fprintf(f, "\t\tgopapageno.%s,\n", opts.Strategy)
	fmt.Fprintf(f, "\t\topts...)\n}\n\n")
}

func (p *parserDescriptor) emitParserFunctions(f io.Writer) {
	fmt.Fprintf(f, "\tfn := func(rule uint16, lhs *gopapageno.Token, rhs []*gopapageno.Token, thread int){\n")
	fmt.Fprintf(f, "\t\tvar ruleType gopapageno.RuleType\n")
	fmt.Fprintf(f, "\t\tswitch rule {\n")
	for i, rule := range p.rules {
		fmt.Fprintf(f, "\t\tcase %d:\n", i)
		fmt.Fprintf(f, "\t\t\truleType = gopapageno.%s\n\n", rule.Type)
		fmt.Fprintf(f, "\t\t\t%s0 := lhs\n", rule.LHS)
		for j, _ := range rule.RHS {
			fmt.Fprintf(f, "\t\t\t%s%d := rhs[%d]\n", rule.RHS[j], j+1, j)
		}
		fmt.Fprintf(f, "\n")

		switch rule.Type {
		case gopapageno.RuleSimple, gopapageno.RuleCyclic:
			if len(rule.RHS) > 0 {
				fmt.Fprintf(f, "\t\t\t%s0.Child = %s1\n", rule.LHS, rule.RHS[0])
				for j := 0; j < len(rule.RHS)-1; j++ {
					fmt.Fprintf(f, "\t\t\t%s%d.Next = %s%d\n", rule.RHS[j], j+1, rule.RHS[j+1], j+2)
				}
				fmt.Fprintf(f, "\t\t\t%s0.LastChild = %s%d\n", rule.LHS, rule.RHS[len(rule.RHS)-1], len(rule.RHS))
			}
		case gopapageno.RuleAppendLeft:
			fmt.Fprintf(f, "\t\t\toldChild := %s0\n", rule.LHS)
			fmt.Fprintf(f, "\t\t\t%s0.Child = %s1\n", rule.LHS, rule.RHS[0])
			for j := 0; j < len(rule.RHS)-1; j++ {
				fmt.Fprintf(f, "\t\t\t%s%d.Next = %s%d\n", rule.RHS[j], j+1, rule.RHS[j+1], j+2)
			}
			fmt.Fprintf(f, "\t\t\t%s%d.Next = oldChild\n", rule.RHS[len(rule.RHS)-1], len(rule.RHS))
		case gopapageno.RuleAppendRight:
			//fmt.Fprintf(f, "\t\t\tlastChild := %s0.Child\n", rule.LHS)
			//fmt.Fprintf(f, "\t\t\tfor t := lastChild.Next; t != nil; t = t.Next {\n")
			//fmt.Fprintf(f, "\t\t\t\tlastChild = t\n\t\t\t}\n\n")
			//fmt.Fprintf(f, "\t\t\tlastChild.Next = %s2\n", rule.RHS[1])
			fmt.Fprintf(f, "\t\t\t%s0.LastChild.Next = %s2\n", rule.LHS, rule.RHS[1])

			for j := 1; j < len(rule.RHS)-1; j++ {
				fmt.Fprintf(f, "\t\t\t%s%d.Next = %s%d\n", rule.RHS[j], j+1, rule.RHS[j+1], j+2)
			}

			fmt.Fprintf(f, "\t\t\t%s0.LastChild = %s%d\n", rule.LHS, rule.RHS[len(rule.RHS)-1], len(rule.RHS))
		case gopapageno.RuleCombine:
			//fmt.Fprintf(f, "\t\t\tlastChild := %s0.Child\n", rule.LHS)
			//fmt.Fprintf(f, "\t\t\tfor t := lastChild.Next; t != nil; t = t.Next {\n")
			//fmt.Fprintf(f, "\t\t\t\tlastChild = t\n\t\t\t}\n\n")
			//fmt.Fprintf(f, "\t\t\tlastChild.Next = %s2\n", rule.RHS[1])
			fmt.Fprintf(f, "\t\t\t%s0.LastChild.Next = %s2\n", rule.LHS, rule.RHS[1])

			l := len(rule.RHS)
			for j := 1; j < l-2; j++ {
				fmt.Fprintf(f, "\t\t\t%s%d.Next = %s%d\n", rule.RHS[j], j+1, rule.RHS[j+1], j+2)
			}

			fmt.Fprintf(f, "\t\t\t%s%d.Next = %s%d.Child\n", rule.RHS[l-2], l-1, rule.RHS[l-1], l)

			fmt.Fprintf(f, "\t\t\t%s0.LastChild = %s%d.LastChild\n", rule.LHS, rule.RHS[len(rule.RHS)-1], len(rule.RHS))
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

		for j, _ := range rule.RHS {
			fmt.Fprintf(f, "\t\t\t_ = %s%d\n", rule.RHS[j], j+1)
		}
	}
	fmt.Fprintf(f, "\t\t}\n")
	fmt.Fprintf(f, "\t\t_ = ruleType\n")
	fmt.Fprintf(f, "\t}\n\n")
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

	fmt.Fprintf(f, "func SprintToken[TokenValue any](root *gopapageno.Token) string {\n")
	fmt.Fprintf(f, "\tvar sprintRec func(t *gopapageno.Token, sb *strings.Builder, indent string)\n\n")
	fmt.Fprintf(f, "\tsprintRec = func(t *gopapageno.Token, sb *strings.Builder, indent string) {\n\t\t")
	fmt.Fprintf(f, `if t == nil {
			return
		}

		sb.WriteString(indent)
		if t.Next == nil {
			sb.WriteString("└── ")
			indent += "    "
		} else {
			sb.WriteString("├── ")
			indent += "|   "
		}
`)

	fmt.Fprintf(f, "\n\t\tswitch t.Type {\n")

	for _, token := range p.nonterminals.Slice() {
		if token == "_EMPTY" {
			fmt.Fprintf(f, "\t\tcase gopapageno.TokenEmpty:\n\t\t\tsb.WriteString(\"Empty\")\n")
		} else {
			fmt.Fprintf(f, "\t\tcase %s:\n\t\t\tsb.WriteString(\"%s\")\n", token, token)
		}
	}

	for _, token := range p.terminals.Slice() {
		if token == "_TERM" {
			fmt.Fprintf(f, "\t\tcase gopapageno.TokenTerm:\n\t\t\tsb.WriteString(\"Term\")\n")
		} else {
			fmt.Fprintf(f, "\t\tcase %s:\n\t\t\tsb.WriteString(\"%s\")\n", token, token)
		}
	}

	fmt.Fprintf(f, "\t\tdefault:\n\t\t\tsb.WriteString(\"Unknown\")\n\t\t}\n")
	fmt.Fprintf(f, "\t\tif t.Value != nil {\n\t\t\tsb.WriteString(fmt.Sprintf(\": %%v\", *t.Value.(*TokenValue)))\n\t\t}\n")
	fmt.Fprintf(f, "\t\tsb.WriteString(\"\\n\")\n\n")

	fmt.Fprintf(f, `		sprintRec(t.Child, sb, indent)
		sprintRec(t.Next, sb, indent[:len(indent)-4])
	}
	`)

	fmt.Fprintf(f, `
	var sb strings.Builder
	
	sprintRec(root, &sb, "")
	
	return sb.String()
}
`)
}
