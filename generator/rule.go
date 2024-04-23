package generator

import (
	"fmt"
	"slices"
	"strings"
)

type rule struct {
	LHS    string
	RHS    []string
	Action string
}

func (r rule) String() string {
	return fmt.Sprintf("%s -> %s", r.LHS, strings.Join(r.RHS, " "))
}

func Productions(rules []rule, index int) [][]string {
	rule := rules[index]

	productions := make([][]string, 1)
	productions[0] = rule.RHS

	for i, t := range rule.RHS {
		for _, r := range rules {
			if t != r.LHS {
				continue
			}

			production := slices.Concat(rule.RHS[:i], r.RHS, rule.RHS[i+1:])
			productions = append(productions, production)
		}
	}

	return productions
}
func RuleProductions(rules []rule, index int) [][]string {
	rule := rules[index]

	productions := make([][]string, 1)
	productions[0] = rule.RHS

	for i, t := range rule.RHS {
		if t != rule.LHS {
			continue
		}

		production := slices.Concat(rule.RHS[:i], rule.RHS, rule.RHS[i+1:])
		productions = append(productions, production)
	}

	return productions
}
