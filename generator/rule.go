package generator

import (
	"fmt"
	"github.com/giornetta/gopapageno"
	"strings"
)

type rule struct {
	LHS    string
	RHS    []string
	Action string
	Type   gopapageno.RuleType
}

func (r rule) String() string {
	return fmt.Sprintf("%s -> %s", r.LHS, strings.Join(r.RHS, " "))
}
