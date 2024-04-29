package generator

import (
	"strings"
)

func (p *parserDescriptor) deleteCopyRules(rulesDict *rulesDictionary) {
	copySets := make(map[string]*set[string], p.nonterminals.Len())

	rhsDict := make(map[string][][]string)

	for _, nonterminal := range p.nonterminals.Iter {
		copySets[nonterminal] = newSet[string]()
	}

	for _, rule := range p.rules {
		// If the rule produces a single nonterminal token,
		// add it to the copy rules set and remove it from the rules' dictionary.
		// Otherwise, add the rule to rhsRule.
		if len(rule.RHS) == 1 && p.nonterminals.Contains(rule.RHS[0]) {
			copySets[rule.LHS].Add(rule.RHS[0])
			rulesDict.Remove(rule.RHS)
		} else {
			if _, ok := rhsDict[rule.LHS]; ok {
				rhsDict[rule.LHS] = append(rhsDict[rule.LHS], rule.RHS)
			} else {
				rhsDict[rule.LHS] = [][]string{rule.RHS}
			}
		}
	}

	changedCopySets := true
	for changedCopySets {
		changedCopySets = false

		for _, nonterminal := range p.nonterminals.Iter {
			lenCopySet := copySets[nonterminal].Len()

			iterCopy := copySets[nonterminal].Copy()

			for _, copyRhs := range iterCopy.Iter {
				for _, curNonterm := range copySets[copyRhs].Iter {
					copySets[nonterminal].Add(curNonterm)
				}
			}

			if lenCopySet < copySets[nonterminal].Len() {
				changedCopySets = true
			}
		}
	}

	for _, nonterminal := range p.nonterminals.Iter {
		for _, curCopyRHS := range copySets[nonterminal].Iter {
			rhsDictCopyRHSs := rhsDict[curCopyRHS]
			for _, rhs := range rhsDictCopyRHSs {
				// There's no need to specify semantic actions
				// because they are already linked to the proper rhs
				rulesDict.Add(&rule{
					LHS:    nonterminal,
					RHS:    rhs,
					Action: "",
				})
			}
		}
	}
}

func (p *parserDescriptor) deleteRepeatedRHS() {
	newRules := make([]rule, 0)

	// Create a rules dictionary and add every parsed rule to it.
	dictRules := newRulesDictionary(len(p.rules))
	for _, rule := range p.rules {
		dictRules.Add(&rule)
	}

	// Delete copy rules from the dictionary
	// TODO: Add explanation of why this is used from the Papers.
	p.deleteCopyRules(dictRules)

	// Create a dictionary that will contain the newly added rules.
	newRulesDict := newRulesDictionary(dictRules.Len())

	// Range over the current rules, check if the RHS contains any nonterminal
	// If it doesn't (i.e. it is a *terminal rule*), add it to the new rules dictionary and remove it from the old dict.
	copyDictRules := dictRules.Copy()
	for i, _ := range copyDictRules.KeysRHS {
		keyRHS := copyDictRules.KeysRHS[i]
		valueLHS := copyDictRules.ValuesLHS[i]
		semAction := copyDictRules.SemActions[i]

		isTerminalRule := true
		for _, token := range keyRHS {
			if p.nonterminals.Contains(token) {
				isTerminalRule = false
				break
			}
		}

		if isTerminalRule {
			for _, curLHS := range valueLHS.Iter {
				newRulesDict.Add(&rule{
					LHS:    curLHS,
					RHS:    keyRHS,
					Action: *semAction,
				})
			}

			dictRules.Remove(keyRHS)
		}
	}

	V := dictRules.LHSSets()

	// Replace token names in prefixes
	newPrefixes := make([][]string, 0)
	for _, prefix := range p.prefixes {
		newPrefixes = append(newPrefixes, p.replaceTokenNames(prefix, V)...)
	}
	p.prefixes = newPrefixes

	dictRulesForIteration := newRulesDictionary(0)
	loop := true
	for loop {
		// Substitutes RHS keys with newly formatted ones.
		for i, _ := range dictRules.KeysRHS {
			keyRHS := dictRules.KeysRHS[i]
			valueLHS := dictRules.ValuesLHS[i]
			semAction := dictRules.SemActions[i]

			addNewRules(dictRulesForIteration, p, keyRHS, valueLHS, semAction, V)
		}

		valueLHSSets := dictRulesForIteration.LHSSets()
		addedNonterminals := make([]*set[string], 0)

		for _, curNontermSet := range valueLHSSets {
			contained := false
			for _, otherNonTermSet := range V {
				if curNontermSet.Equals(otherNonTermSet) {
					contained = true
					break
				}
			}
			if !contained {
				addedNonterminals = append(addedNonterminals, curNontermSet)
				V = append(V, curNontermSet)
			}
		}

		for i, _ := range dictRulesForIteration.KeysRHS {
			keyRHS := dictRulesForIteration.KeysRHS[i]
			valueLHS := dictRulesForIteration.ValuesLHS[i]
			semAction := dictRulesForIteration.SemActions[i]

			for _, curLHS := range valueLHS.Iter {
				newRulesDict.Add(&rule{
					LHS:    curLHS,
					RHS:    keyRHS,
					Action: *semAction,
				})
			}
		}

		if len(addedNonterminals) == 0 {
			loop = false
		}
	}

	// TODO: remove unused nonterminals (see cpapageno)

	newAxiom := "NEW_AXIOM"
	newAxiomSet := newSet[string]()
	newAxiomSet.Add(newAxiom)

	newAxiomSemAction := "{\n\t$$.Value = $1.Value\n}"

	V = append(V, newAxiomSet)

	for _, nontermSet := range V {
		if nontermSet.Contains(p.axiom) {
			newRulesDict.Add(&rule{
				LHS:    newAxiom,
				RHS:    []string{strings.Join(nontermSet.Slice(), "_")},
				Action: newAxiomSemAction,
			})
		}
	}

	// Create the rules from rulesDictionary
	for i, _ := range newRulesDict.KeysRHS {
		keyRHS := newRulesDict.KeysRHS[i]
		valueLHS := newRulesDict.ValuesLHS[i]
		semAction := newRulesDict.SemActions[i]

		newRules = append(newRules, rule{strings.Join(valueLHS.Slice(), "_"), keyRHS, *semAction})
	}

	p.rules = newRules
	p.inferTokens()

	p.axiom = newAxiom
}

func (p *parserDescriptor) replaceTokenNames(keyRHS []string, newNonterminals []*set[string]) [][]string {
	newTokenNames := make([][]string, 0)

	var rec func(tokens []string, newTokens []string)
	rec = func(tokens []string, newTokens []string) {
		if len(tokens) == 0 {
			newTokenNames = append(newTokenNames, newTokens)
			return
		}

		token := tokens[0]
		if p.nonterminals.Contains(token) {
			for _, nonTermSuperSet := range newNonterminals {
				if nonTermSuperSet.Contains(token) {
					newTokens = append(newTokens, strings.Join(nonTermSuperSet.Slice(), "_"))
					rec(tokens[1:], newTokens)

					newTokensCopy := make([]string, len(newTokens)-1)
					copy(newTokensCopy, newTokens)
					newTokens = newTokensCopy
				}
			}
		} else {
			newTokens = append(newTokens, token)
			rec(tokens[1:], newTokens)

			newTokensCopy := make([]string, len(newTokens)-1)
			copy(newTokensCopy, newTokens)
			newTokens = newTokensCopy
		}
	}

	newTokens := make([]string, 0)
	rec(keyRHS, newTokens)

	return newTokenNames
}

func addNewRules(dict *rulesDictionary, p *parserDescriptor,
	keyRHS []string, valueLHS *set[string], semAction *string, newNonterminals []*set[string]) {

	newRuleRHS := p.replaceTokenNames(keyRHS, newNonterminals)

	for _, lhs := range valueLHS.Iter {
		for _, rhs := range newRuleRHS {
			dict.Add(&rule{
				LHS:    lhs,
				RHS:    rhs,
				Action: *semAction,
			})
		}

	}

	return
}

func (p *parserDescriptor) sortRulesByRHS() {
	sortedRules := make([]rule, 0, len(p.rules))

	for _, curRule := range p.rules {
		insertPosition := -1
		for i, curSortedRule := range sortedRules {
			if p.rhsLessThan(curRule.RHS, curSortedRule.RHS) {
				insertPosition = i
				break
			}
		}
		if insertPosition == -1 {
			sortedRules = append(sortedRules, curRule)
		} else {
			sortedRules = append(sortedRules, rule{})
			copy(sortedRules[insertPosition+1:], sortedRules[insertPosition:])
			sortedRules[insertPosition] = curRule
		}
	}

	p.rules = sortedRules
}

func rhsEquals(rhs1 []string, rhs2 []string) bool {
	if len(rhs1) != len(rhs2) {
		return false
	}

	for i, _ := range rhs1 {
		if rhs1[i] != rhs2[i] {
			return false
		}
	}

	return true
}

func (p *parserDescriptor) rhsLessThan(rhs1 []string, rhs2 []string) bool {
	minLen := len(rhs1)
	if len(rhs2) < minLen {
		minLen = len(rhs2)
	}

	for i := 0; i < minLen; i++ {
		//If the first is in nonterminals and the second is in terminals,
		//the first token is certainly less than the second
		if p.nonterminals.Contains(rhs1[i]) && p.terminals.Contains(rhs2[i]) {
			return true
		}
		//If the first is in terminals and the second is in nonterminals,
		//the first token is certainly greater than the second
		if p.terminals.Contains(rhs1[i]) && p.nonterminals.Contains(rhs2[i]) {
			return false
		}

		if rhs1[i] < rhs2[i] {
			return true
		}
		if rhs1[i] > rhs2[i] {
			return false
		}
	}

	if len(rhs1) < len(rhs2) {
		return true
	}
	return false
}
