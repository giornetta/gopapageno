// Code generated by Gopapageno; DO NOT EDIT.
package main

import (
	"fmt"
	"github.com/giornetta/gopapageno"
	"strings"
)

func ParserPreallocMem(inputSize int, numThreads int) {
}

// Non-terminals
const (
	Array_Elements_Value = gopapageno.TokenEmpty + 1 + iota
	Document_Elements_Object_Value
	Elements
	Elements_String_Value
	Members
	NEW_AXIOM
)

// Terminals
const (
	COMMA = gopapageno.TokenTerm + 1 + iota
	LCURLY
	LSQUARE
	QUOTE
	RCURLY
	RSQUARE
	STRING
)

func SprintToken[TokenValue any](root *gopapageno.Token) string {
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
		case Array_Elements_Value:
			sb.WriteString("Array_Elements_Value")
		case Document_Elements_Object_Value:
			sb.WriteString("Document_Elements_Object_Value")
		case Elements:
			sb.WriteString("Elements")
		case Elements_String_Value:
			sb.WriteString("Elements_String_Value")
		case Members:
			sb.WriteString("Members")
		case NEW_AXIOM:
			sb.WriteString("NEW_AXIOM")
		case gopapageno.TokenEmpty:
			sb.WriteString("Empty")
		case COMMA:
			sb.WriteString("COMMA")
		case LCURLY:
			sb.WriteString("LCURLY")
		case LSQUARE:
			sb.WriteString("LSQUARE")
		case QUOTE:
			sb.WriteString("QUOTE")
		case RCURLY:
			sb.WriteString("RCURLY")
		case RSQUARE:
			sb.WriteString("RSQUARE")
		case STRING:
			sb.WriteString("STRING")
		case gopapageno.TokenTerm:
			sb.WriteString("Term")
		default:
			sb.WriteString("Unknown")
		}
		if t.Value != nil {
			sb.WriteString(fmt.Sprintf(": %v", *t.Value.(*TokenValue)))
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
	numTerminals := uint16(8)
	numNonTerminals := uint16(7)

	maxRHSLen := 3
	rules := []gopapageno.Rule{
		{Elements, []gopapageno.TokenType{Array_Elements_Value, COMMA, Array_Elements_Value}},
		{Elements, []gopapageno.TokenType{Array_Elements_Value, COMMA, Document_Elements_Object_Value}},
		{Elements, []gopapageno.TokenType{Array_Elements_Value, COMMA, Elements}},
		{NEW_AXIOM, []gopapageno.TokenType{Document_Elements_Object_Value}},
		{Elements, []gopapageno.TokenType{Document_Elements_Object_Value, COMMA, Array_Elements_Value}},
		{Elements, []gopapageno.TokenType{Document_Elements_Object_Value, COMMA, Document_Elements_Object_Value}},
		{Elements, []gopapageno.TokenType{Document_Elements_Object_Value, COMMA, Elements}},
		{Elements, []gopapageno.TokenType{Elements, COMMA, Array_Elements_Value}},
		{Elements, []gopapageno.TokenType{Elements, COMMA, Document_Elements_Object_Value}},
		{Elements, []gopapageno.TokenType{Elements, COMMA, Elements}},
		{Members, []gopapageno.TokenType{Members, COMMA, Members}},
		{Document_Elements_Object_Value, []gopapageno.TokenType{LCURLY, Members, RCURLY}},
		{Document_Elements_Object_Value, []gopapageno.TokenType{LCURLY, RCURLY}},
		{Array_Elements_Value, []gopapageno.TokenType{LSQUARE, Array_Elements_Value, RSQUARE}},
		{Array_Elements_Value, []gopapageno.TokenType{LSQUARE, Document_Elements_Object_Value, RSQUARE}},
		{Array_Elements_Value, []gopapageno.TokenType{LSQUARE, Elements, RSQUARE}},
		{Array_Elements_Value, []gopapageno.TokenType{LSQUARE, RSQUARE}},
		{Elements_String_Value, []gopapageno.TokenType{QUOTE, QUOTE}},
		{Elements_String_Value, []gopapageno.TokenType{QUOTE, STRING, QUOTE}},
	}
	compressedRules := []uint16{0, 0, 7, 1, 17, 2, 40, 3, 63, 5, 86, 32770, 99, 32771, 117, 32772, 155, 0, 0, 1, 32769, 22, 0, 0, 3, 1, 31, 2, 34, 3, 37, 3, 0, 0, 3, 1, 0, 3, 2, 0, 6, 3, 1, 32769, 45, 0, 0, 3, 1, 54, 2, 57, 3, 60, 3, 4, 0, 3, 5, 0, 3, 6, 0, 0, 0, 1, 32769, 68, 0, 0, 3, 1, 77, 2, 80, 3, 83, 3, 7, 0, 3, 8, 0, 3, 9, 0, 0, 0, 1, 32769, 91, 0, 0, 1, 5, 96, 5, 10, 0, 0, 0, 2, 5, 106, 32773, 114, 0, 0, 1, 32773, 111, 2, 11, 0, 2, 12, 0, 0, 0, 4, 1, 128, 2, 136, 3, 144, 32774, 152, 0, 0, 1, 32774, 133, 1, 13, 0, 0, 0, 1, 32774, 141, 1, 14, 0, 0, 0, 1, 32774, 149, 1, 15, 0, 1, 16, 0, 0, 0, 2, 32772, 162, 32775, 165, 4, 17, 0, 0, 0, 1, 32772, 170, 4, 18, 0}

	maxPrefixLen := 0
	prefixes := [][]gopapageno.TokenType{}
	precMatrix := [][]gopapageno.Precedence{
		{gopapageno.PrecEquals, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields},
		{gopapageno.PrecTakes, gopapageno.PrecAssociative, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecEquals, gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecTakes, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecTakes, gopapageno.PrecEquals},
		{gopapageno.PrecTakes, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals, gopapageno.PrecEquals},
	}
	bitPackedMatrix := []uint64{
		24206874444191060, 598177812709378,
	}

	fn := func(rule uint16, lhs *gopapageno.Token, rhs []*gopapageno.Token, thread int) {
		switch rule {
		case 0:
			Elements0 := lhs
			Array_Elements_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Array_Elements_Value3 := rhs[2]

			Elements0.Child = Array_Elements_Value1
			Array_Elements_Value1.Next = COMMA2
			COMMA2.Next = Array_Elements_Value3

			{
			}
		case 1:
			Elements0 := lhs
			Array_Elements_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Document_Elements_Object_Value3 := rhs[2]

			Elements0.Child = Array_Elements_Value1
			Array_Elements_Value1.Next = COMMA2
			COMMA2.Next = Document_Elements_Object_Value3

			{
			}
		case 2:
			Elements0 := lhs
			Array_Elements_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Elements3 := rhs[2]

			Elements0.Child = Array_Elements_Value1
			Array_Elements_Value1.Next = COMMA2
			COMMA2.Next = Elements3

			{
			}
		case 3:
			NEW_AXIOM0 := lhs
			Document_Elements_Object_Value1 := rhs[0]

			NEW_AXIOM0.Child = Document_Elements_Object_Value1

			{
				NEW_AXIOM0.Value = Document_Elements_Object_Value1.Value
			}
		case 4:
			Elements0 := lhs
			Document_Elements_Object_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Array_Elements_Value3 := rhs[2]

			Elements0.Child = Document_Elements_Object_Value1
			Document_Elements_Object_Value1.Next = COMMA2
			COMMA2.Next = Array_Elements_Value3

			{
			}
		case 5:
			Elements0 := lhs
			Document_Elements_Object_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Document_Elements_Object_Value3 := rhs[2]

			Elements0.Child = Document_Elements_Object_Value1
			Document_Elements_Object_Value1.Next = COMMA2
			COMMA2.Next = Document_Elements_Object_Value3

			{
			}
		case 6:
			Elements0 := lhs
			Document_Elements_Object_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Elements3 := rhs[2]

			Elements0.Child = Document_Elements_Object_Value1
			Document_Elements_Object_Value1.Next = COMMA2
			COMMA2.Next = Elements3

			{
			}
		case 7:
			Elements0 := lhs
			Elements1 := rhs[0]
			COMMA2 := rhs[1]
			Array_Elements_Value3 := rhs[2]

			Elements0.Child = Elements1
			Elements1.Next = COMMA2
			COMMA2.Next = Array_Elements_Value3

			{
			}
		case 8:
			Elements0 := lhs
			Elements1 := rhs[0]
			COMMA2 := rhs[1]
			Document_Elements_Object_Value3 := rhs[2]

			Elements0.Child = Elements1
			Elements1.Next = COMMA2
			COMMA2.Next = Document_Elements_Object_Value3

			{
			}
		case 9:
			Elements0 := lhs
			Elements1 := rhs[0]
			COMMA2 := rhs[1]
			Elements3 := rhs[2]

			Elements0.Child = Elements1
			Elements1.Next = COMMA2
			COMMA2.Next = Elements3

			{
			}
		case 10:
			Members0 := lhs
			Members1 := rhs[0]
			COMMA2 := rhs[1]
			Members3 := rhs[2]

			Members0.Child = Members1
			Members1.Next = COMMA2
			COMMA2.Next = Members3

			{
			}
		case 11:
			Document_Elements_Object_Value0 := lhs
			LCURLY1 := rhs[0]
			Members2 := rhs[1]
			RCURLY3 := rhs[2]

			Document_Elements_Object_Value0.Child = LCURLY1
			LCURLY1.Next = Members2
			Members2.Next = RCURLY3

			{
			}
		case 12:
			Document_Elements_Object_Value0 := lhs
			LCURLY1 := rhs[0]
			RCURLY2 := rhs[1]

			Document_Elements_Object_Value0.Child = LCURLY1
			LCURLY1.Next = RCURLY2

			{
			}
		case 13:
			Array_Elements_Value0 := lhs
			LSQUARE1 := rhs[0]
			Array_Elements_Value2 := rhs[1]
			RSQUARE3 := rhs[2]

			Array_Elements_Value0.Child = LSQUARE1
			LSQUARE1.Next = Array_Elements_Value2
			Array_Elements_Value2.Next = RSQUARE3

			{
			}
		case 14:
			Array_Elements_Value0 := lhs
			LSQUARE1 := rhs[0]
			Document_Elements_Object_Value2 := rhs[1]
			RSQUARE3 := rhs[2]

			Array_Elements_Value0.Child = LSQUARE1
			LSQUARE1.Next = Document_Elements_Object_Value2
			Document_Elements_Object_Value2.Next = RSQUARE3

			{
			}
		case 15:
			Array_Elements_Value0 := lhs
			LSQUARE1 := rhs[0]
			Elements2 := rhs[1]
			RSQUARE3 := rhs[2]

			Array_Elements_Value0.Child = LSQUARE1
			LSQUARE1.Next = Elements2
			Elements2.Next = RSQUARE3

			{
			}
		case 16:
			Array_Elements_Value0 := lhs
			LSQUARE1 := rhs[0]
			RSQUARE2 := rhs[1]

			Array_Elements_Value0.Child = LSQUARE1
			LSQUARE1.Next = RSQUARE2

			{
			}
		case 17:
			Elements_String_Value0 := lhs
			QUOTE1 := rhs[0]
			QUOTE2 := rhs[1]

			Elements_String_Value0.Child = QUOTE1
			QUOTE1.Next = QUOTE2

			{
			}
		case 18:
			Elements_String_Value0 := lhs
			QUOTE1 := rhs[0]
			STRING2 := rhs[1]
			QUOTE3 := rhs[2]

			Elements_String_Value0.Child = QUOTE1
			QUOTE1.Next = STRING2
			STRING2.Next = QUOTE3

			{
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
		prefixes,
		maxPrefixLen,
		precMatrix,
		bitPackedMatrix,
		fn,
		gopapageno.AOPP,
		opts...)
}