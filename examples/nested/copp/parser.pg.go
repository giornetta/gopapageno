// Code generated by Gopapageno; DO NOT EDIT.
package main

import (
	"github.com/giornetta/gopapageno"
	"strings"
	"fmt"
)


// Non-terminals
const (
	E = gopapageno.TokenEmpty + 1 + iota
	S
	V
)

// Terminals
const (
	NUMBER = gopapageno.TokenTerm + 1 + iota
	OPERATOR
	TIMES
)

func SprintToken[ValueType any](root *gopapageno.Token) string {
	var sprintRec func(t *gopapageno.Token, sb *strings.Builder, indent string)

	sprintRec = func(t *gopapageno.Token, sb *strings.Builder, indent string) {
		if t == nil {
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

		switch t.Type {
		case E:
			sb.WriteString("E")
		case S:
			sb.WriteString("S")
		case V:
			sb.WriteString("V")
		case gopapageno.TokenEmpty:
			sb.WriteString("Empty")
		case NUMBER:
			sb.WriteString("NUMBER")
		case OPERATOR:
			sb.WriteString("OPERATOR")
		case TIMES:
			sb.WriteString("TIMES")
		case gopapageno.TokenTerm:
			sb.WriteString("Term")
		default:
			sb.WriteString("Unknown")
		}

		if t.Value != nil {
			if v, ok := any(t.Value).(*ValueType); ok {
				sb.WriteString(fmt.Sprintf(": %v", *v))
			}
		}
		
		sb.WriteString("\n")
		
		sprintRec(t.Child, sb, indent)
		sprintRec(t.Next, sb, indent[:len(indent)-4])
	}

	var sb strings.Builder
	
	sprintRec(root, &sb, "")
	
	return sb.String()
}

func NewGrammar() *gopapageno.Grammar {
	numTerminals := uint16(4)
	numNonTerminals := uint16(4)

	maxRHSLen := 4
	rules := []gopapageno.Rule{
		{S, []gopapageno.TokenType{E}, gopapageno.RuleSimple},
		{E, []gopapageno.TokenType{V, TIMES, OPERATOR, V}, gopapageno.RuleCyclic},
		{V, []gopapageno.TokenType{NUMBER}, gopapageno.RuleSimple},
	}
	compressedRules := []uint16{0, 0, 3, 1, 9, 3, 12, 32769, 30, 2, 0, 0, 0, 0, 1, 32771, 17, 0, 0, 1, 32770, 22, 0, 0, 1, 3, 27, 1, 1, 0, 3, 2, 0	}

	maxPrefixLength := 6
	prefixes := [][]gopapageno.TokenType{
		{V, TIMES, OPERATOR, V, TIMES, OPERATOR},
	}
	compressedPrefixes := []uint16{0, 0, 1, 3, 5, 0, 0, 1, 32771, 10, 0, 0, 1, 32770, 15, 0, 0, 1, 3, 20, 0, 0, 1, 32771, 25, 0, 0, 1, 32770, 30, 1, 1, 0	}

	precMatrix := [][]gopapageno.Precedence{
		{gopapageno.PrecEquals, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields},
		{gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecEquals, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEquals},
	}
	bitPackedMatrix := []uint64{
		33981012, 
	}

	fn := func(ruleDescription uint16, ruleFlags gopapageno.RuleFlags, lhs *gopapageno.Token, rhs []*gopapageno.Token, thread int){
		switch ruleDescription {
		case 0:
			S0 := lhs
			E1 := rhs[0]

			S0.Child = E1
			S0.LastChild = E1

			{
			    S0.Value = E1.Value
			}
			_ = E1
		case 1:
			E0 := lhs
			V1 := rhs[0]
			TIMES2 := rhs[1]
			OPERATOR3 := rhs[2]
			V4 := rhs[3]

			if ruleFlags.Has(gopapageno.RuleAppend) {
				E0.LastChild.Next = TIMES2
			} else {
				E0.Child = V1
				V1.Next = TIMES2
			}

			TIMES2.Next = OPERATOR3
			OPERATOR3.Next = V4

			if ruleFlags.Has(gopapageno.RuleCombine) {
				OPERATOR3.Next = V4.Child
				E0.LastChild = V4.LastChild
			} else {
				OPERATOR3.Next = V4
				E0.LastChild = V4
			}

			{
			    v1 := V1.Value.(int64)
			    v2 := V4.Value.(int64)
			
			    if ruleFlags.Has(gopapageno.RuleAppend) || ruleFlags.Has(gopapageno.RuleCombine) {
			        E0.Value = E0.Value.(int64) + v2
			    } else {
			        E0.Value = v1 + v2
			    }
			}
			_ = V1
			_ = TIMES2
			_ = OPERATOR3
			_ = V4
		case 2:
			V0 := lhs
			NUMBER1 := rhs[0]

			V0.Child = NUMBER1
			V0.LastChild = NUMBER1

			{
			    V0.Value = NUMBER1.Value
			}
			_ = NUMBER1
		}
		_ = ruleFlags
	}

	return &gopapageno.Grammar{
		NumTerminals:  numTerminals,
		NumNonterminals: numNonTerminals,
		MaxRHSLength: maxRHSLen,
		Rules: rules,
		CompressedRules: compressedRules,
		PrecedenceMatrix: precMatrix,
		BitPackedPrecedenceMatrix: bitPackedMatrix,
		MaxPrefixLength: maxPrefixLength,
		Prefixes: prefixes,
		CompressedPrefixes: compressedPrefixes,
		Func: fn,
		ParsingStrategy: gopapageno.COPP,
	}
}

