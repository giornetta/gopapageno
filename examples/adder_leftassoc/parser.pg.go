// Code generated by Gopapageno; DO NOT EDIT.
package main

import (
	"fmt"
	"github.com/giornetta/gopapageno"
	"strings"
)

import (
	"math"
)

var parserPools []*gopapageno.Pool[int64]

func ParserPreallocMem(inputSize int, numThreads int) {
	parserPools = make([]*gopapageno.Pool[int64], numThreads)

	avgCharsPerNumber := float64(2)
	poolSizePerThread := int(math.Ceil((float64(inputSize) / avgCharsPerNumber) / float64(numThreads)))

	for i := 0; i < numThreads; i++ {
		parserPools[i] = gopapageno.NewPool[int64](poolSizePerThread)

	}
}

// Non-terminals
const (
	E_S = gopapageno.TokenEmpty + 1 + iota
	E_S_T
	NEW_AXIOM
)

// Terminals
const (
	LPAR = gopapageno.TokenTerm + 1 + iota
	NUMBER
	PLUS
	RPAR
)

func SprintToken[T any](root *gopapageno.Token) string {
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
		case E_S:
			sb.WriteString("E_S")
		case E_S_T:
			sb.WriteString("E_S_T")
		case NEW_AXIOM:
			sb.WriteString("NEW_AXIOM")
		case gopapageno.TokenEmpty:
			sb.WriteString("Empty")
		case LPAR:
			sb.WriteString("LPAR")
		case NUMBER:
			sb.WriteString("NUMBER")
		case PLUS:
			sb.WriteString("PLUS")
		case RPAR:
			sb.WriteString("RPAR")
		case gopapageno.TokenTerm:
			sb.WriteString("Term")
		default:
			sb.WriteString("Unknown")
		}
		if t.Value != nil {
			sb.WriteString(fmt.Sprintf(": %v", *t.Value.(*T)))
		}
		sb.WriteString("\n")

		sprintRec(t.Child, sb, indent)
		sprintRec(t.Next, sb, indent[:len(indent)-4])
	}

	var sb strings.Builder

	sprintRec(root, &sb, "")

	return sb.String()
}

func NewParser(opts ...gopapageno.ParserOpt) *gopapageno.Parser {
	numTerminals := uint16(5)
	numNonTerminals := uint16(4)

	maxRHSLen := 3
	rules := []gopapageno.Rule{
		{NEW_AXIOM, []gopapageno.TokenType{E_S}},
		{E_S, []gopapageno.TokenType{E_S, PLUS, E_S_T}},
		{NEW_AXIOM, []gopapageno.TokenType{E_S_T}},
		{E_S, []gopapageno.TokenType{E_S_T, PLUS, E_S_T}},
		{E_S_T, []gopapageno.TokenType{LPAR, E_S, RPAR}},
		{E_S_T, []gopapageno.TokenType{LPAR, E_S_T, RPAR}},
		{E_S_T, []gopapageno.TokenType{NUMBER}},
	}
	compressedRules := []uint16{0, 0, 4, 1, 11, 2, 24, 32769, 37, 32770, 60, 3, 0, 1, 32771, 16, 0, 0, 1, 2, 21, 1, 1, 0, 3, 2, 1, 32771, 29, 0, 0, 1, 2, 34, 1, 3, 0, 0, 0, 2, 1, 44, 2, 52, 0, 0, 1, 32772, 49, 2, 4, 0, 0, 0, 1, 32772, 57, 2, 5, 0, 2, 6, 0}

	precMatrix := [][]gopapageno.Precedence{
		{gopapageno.PrecEquals, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecTakes},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecTakes, gopapageno.PrecTakes},
		{gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecTakes},
	}
	bitPackedMatrix := []uint64{
		772547241314305,
	}

	fn := func(rule uint16, lhs *gopapageno.Token, rhs []*gopapageno.Token, thread int) {
		switch rule {
		case 0:
			NEW_AXIOM0 := lhs
			E_S1 := rhs[0]

			NEW_AXIOM0.Child = E_S1

			{
				NEW_AXIOM0.Value = E_S1.Value
			}
		case 1:
			E_S0 := lhs
			E_S1 := rhs[0]
			PLUS2 := rhs[1]
			E_S_T3 := rhs[2]

			E_S0.Child = E_S1
			E_S1.Next = PLUS2
			PLUS2.Next = E_S_T3

			{
				newValue := parserPools[thread].Get()
				*newValue = *E_S1.Value.(*int64) + *E_S_T3.Value.(*int64)
				E_S0.Value = newValue
			}
		case 2:
			NEW_AXIOM0 := lhs
			E_S_T1 := rhs[0]

			NEW_AXIOM0.Child = E_S_T1

			{
				NEW_AXIOM0.Value = E_S_T1.Value
			}
		case 3:
			E_S0 := lhs
			E_S_T1 := rhs[0]
			PLUS2 := rhs[1]
			E_S_T3 := rhs[2]

			E_S0.Child = E_S_T1
			E_S_T1.Next = PLUS2
			PLUS2.Next = E_S_T3

			{
				newValue := parserPools[thread].Get()
				*newValue = *E_S_T1.Value.(*int64) + *E_S_T3.Value.(*int64)
				E_S0.Value = newValue
			}
		case 4:
			E_S_T0 := lhs
			LPAR1 := rhs[0]
			E_S2 := rhs[1]
			RPAR3 := rhs[2]

			E_S_T0.Child = LPAR1
			LPAR1.Next = E_S2
			E_S2.Next = RPAR3

			{
				E_S_T0.Value = E_S2.Value
			}
		case 5:
			E_S_T0 := lhs
			LPAR1 := rhs[0]
			E_S_T2 := rhs[1]
			RPAR3 := rhs[2]

			E_S_T0.Child = LPAR1
			LPAR1.Next = E_S_T2
			E_S_T2.Next = RPAR3

			{
				E_S_T0.Value = E_S_T2.Value
			}
		case 6:
			E_S_T0 := lhs
			NUMBER1 := rhs[0]

			E_S_T0.Child = NUMBER1

			{
				E_S_T0.Value = NUMBER1.Value
			}
		}
	}

	return gopapageno.NewParser(
		NewLexer(),
		numTerminals,
		numNonTerminals,
		maxRHSLen,
		rules,
		compressedRules,
		precMatrix,
		bitPackedMatrix,
		fn,
		opts...)
}