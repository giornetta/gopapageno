// Code generated by Gopapageno; DO NOT EDIT.
package main

import (
	"fmt"
	"github.com/giornetta/gopapageno"
	"strings"
)

// Non-terminals
const (
	Array_Elements_Value = gopapageno.TokenEmpty + 1 + iota
	Document
	Elements
	Elements_Object_Value
	Elements_Value
	Members
	Members_Pair
)

// Terminals
const (
	BOOL = gopapageno.TokenTerm + 1 + iota
	COLON
	COMMA
	LCURLY
	LSQUARE
	NULL
	NUMBER
	RCURLY
	RSQUARE
	STRING
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
		case Array_Elements_Value:
			sb.WriteString("Array_Elements_Value")
		case Document:
			sb.WriteString("Document")
		case Elements:
			sb.WriteString("Elements")
		case Elements_Object_Value:
			sb.WriteString("Elements_Object_Value")
		case Elements_Value:
			sb.WriteString("Elements_Value")
		case Members:
			sb.WriteString("Members")
		case Members_Pair:
			sb.WriteString("Members_Pair")
		case gopapageno.TokenEmpty:
			sb.WriteString("Empty")
		case BOOL:
			sb.WriteString("BOOL")
		case COLON:
			sb.WriteString("COLON")
		case COMMA:
			sb.WriteString("COMMA")
		case LCURLY:
			sb.WriteString("LCURLY")
		case LSQUARE:
			sb.WriteString("LSQUARE")
		case NULL:
			sb.WriteString("NULL")
		case NUMBER:
			sb.WriteString("NUMBER")
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
	numTerminals := uint16(11)
	numNonTerminals := uint16(8)

	maxRHSLen := 3
	rules := []gopapageno.Rule{
		{Elements, []gopapageno.TokenType{Array_Elements_Value, COMMA, Array_Elements_Value}, gopapageno.RuleSimple},
		{Elements, []gopapageno.TokenType{Array_Elements_Value, COMMA, Elements_Object_Value}, gopapageno.RuleSimple},
		{Elements, []gopapageno.TokenType{Array_Elements_Value, COMMA, Elements_Value}, gopapageno.RuleSimple},
		{Elements, []gopapageno.TokenType{Elements, COMMA, Array_Elements_Value}, gopapageno.RuleSimple},
		{Elements, []gopapageno.TokenType{Elements, COMMA, Elements_Object_Value}, gopapageno.RuleSimple},
		{Elements, []gopapageno.TokenType{Elements, COMMA, Elements_Value}, gopapageno.RuleSimple},
		{Elements, []gopapageno.TokenType{Elements_Object_Value, COMMA, Array_Elements_Value}, gopapageno.RuleSimple},
		{Elements, []gopapageno.TokenType{Elements_Object_Value, COMMA, Elements_Object_Value}, gopapageno.RuleSimple},
		{Elements, []gopapageno.TokenType{Elements_Object_Value, COMMA, Elements_Value}, gopapageno.RuleSimple},
		{Elements, []gopapageno.TokenType{Elements_Value, COMMA, Array_Elements_Value}, gopapageno.RuleSimple},
		{Elements, []gopapageno.TokenType{Elements_Value, COMMA, Elements_Object_Value}, gopapageno.RuleSimple},
		{Elements, []gopapageno.TokenType{Elements_Value, COMMA, Elements_Value}, gopapageno.RuleSimple},
		{Members, []gopapageno.TokenType{Members, COMMA, Members_Pair}, gopapageno.RuleSimple},
		{Members, []gopapageno.TokenType{Members_Pair, COMMA, Members_Pair}, gopapageno.RuleSimple},
		{Elements_Value, []gopapageno.TokenType{BOOL}, gopapageno.RuleSimple},
		{Elements_Object_Value, []gopapageno.TokenType{LCURLY, Members, RCURLY}, gopapageno.RuleSimple},
		{Elements_Object_Value, []gopapageno.TokenType{LCURLY, Members_Pair, RCURLY}, gopapageno.RuleSimple},
		{Elements_Object_Value, []gopapageno.TokenType{LCURLY, RCURLY}, gopapageno.RuleSimple},
		{Array_Elements_Value, []gopapageno.TokenType{LSQUARE, Array_Elements_Value, RSQUARE}, gopapageno.RuleSimple},
		{Array_Elements_Value, []gopapageno.TokenType{LSQUARE, Elements, RSQUARE}, gopapageno.RuleSimple},
		{Array_Elements_Value, []gopapageno.TokenType{LSQUARE, Elements_Object_Value, RSQUARE}, gopapageno.RuleSimple},
		{Array_Elements_Value, []gopapageno.TokenType{LSQUARE, Elements_Value, RSQUARE}, gopapageno.RuleSimple},
		{Array_Elements_Value, []gopapageno.TokenType{LSQUARE, RSQUARE}, gopapageno.RuleSimple},
		{Elements_Value, []gopapageno.TokenType{NULL}, gopapageno.RuleSimple},
		{Elements_Value, []gopapageno.TokenType{NUMBER}, gopapageno.RuleSimple},
		{Elements_Value, []gopapageno.TokenType{STRING}, gopapageno.RuleSimple},
		{Members_Pair, []gopapageno.TokenType{STRING, COLON, Array_Elements_Value}, gopapageno.RuleSimple},
		{Members_Pair, []gopapageno.TokenType{STRING, COLON, Elements_Object_Value}, gopapageno.RuleSimple},
		{Members_Pair, []gopapageno.TokenType{STRING, COLON, Elements_Value}, gopapageno.RuleSimple},
	}
	compressedRules := []uint16{0, 0, 12, 1, 27, 3, 50, 4, 73, 5, 96, 6, 119, 7, 132, 32769, 145, 32772, 148, 32773, 176, 32774, 224, 32775, 227, 32778, 230, 0, 0, 1, 32771, 32, 0, 0, 3, 1, 41, 4, 44, 5, 47, 3, 0, 0, 3, 1, 0, 3, 2, 0, 0, 0, 1, 32771, 55, 0, 0, 3, 1, 64, 4, 67, 5, 70, 3, 3, 0, 3, 4, 0, 3, 5, 0, 0, 0, 1, 32771, 78, 0, 0, 3, 1, 87, 4, 90, 5, 93, 3, 6, 0, 3, 7, 0, 3, 8, 0, 0, 0, 1, 32771, 101, 0, 0, 3, 1, 110, 4, 113, 5, 116, 3, 9, 0, 3, 10, 0, 3, 11, 0, 0, 0, 1, 32771, 124, 0, 0, 1, 7, 129, 6, 12, 0, 0, 0, 1, 32771, 137, 0, 0, 1, 7, 142, 6, 13, 0, 5, 14, 0, 0, 0, 3, 6, 157, 7, 165, 32776, 173, 0, 0, 1, 32776, 162, 4, 15, 0, 0, 0, 1, 32776, 170, 4, 16, 0, 4, 17, 0, 0, 0, 5, 1, 189, 3, 197, 4, 205, 5, 213, 32777, 221, 0, 0, 1, 32777, 194, 1, 18, 0, 0, 0, 1, 32777, 202, 1, 19, 0, 0, 0, 1, 32777, 210, 1, 20, 0, 0, 0, 1, 32777, 218, 1, 21, 0, 1, 22, 0, 5, 23, 0, 5, 24, 0, 5, 25, 1, 32770, 235, 0, 0, 3, 1, 244, 4, 247, 5, 250, 7, 26, 0, 7, 27, 0, 7, 28, 0}

	precMatrix := [][]gopapageno.Precedence{
		{gopapageno.PrecEquals, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields},
		{gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecEmpty},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecYields},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecYields},
		{gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecYields, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEquals, gopapageno.PrecEmpty, gopapageno.PrecYields},
		{gopapageno.PrecTakes, gopapageno.PrecYields, gopapageno.PrecEmpty, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecYields, gopapageno.PrecEmpty, gopapageno.PrecEquals, gopapageno.PrecYields},
		{gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecEmpty},
		{gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecEmpty},
		{gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecEmpty},
		{gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecEmpty},
		{gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecEmpty, gopapageno.PrecTakes, gopapageno.PrecTakes, gopapageno.PrecEmpty},
	}
	bitPackedMatrix := []uint64{
		2691009079795864916, 1536167278698649113, 36635736172136484, 1301856705847434,
	}

	fn := func(ruleDescription uint16, ruleFlags gopapageno.RuleFlags, lhs *gopapageno.Token, rhs []*gopapageno.Token, thread int) {
		switch ruleDescription {
		case 0:
			Elements0 := lhs
			Array_Elements_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Array_Elements_Value3 := rhs[2]

			Elements0.Child = Array_Elements_Value1
			Array_Elements_Value1.Next = COMMA2
			COMMA2.Next = Array_Elements_Value3
			Elements0.LastChild = Array_Elements_Value3

			{
			}
			_ = Array_Elements_Value1
			_ = COMMA2
			_ = Array_Elements_Value3
		case 1:
			Elements0 := lhs
			Array_Elements_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Elements_Object_Value3 := rhs[2]

			Elements0.Child = Array_Elements_Value1
			Array_Elements_Value1.Next = COMMA2
			COMMA2.Next = Elements_Object_Value3
			Elements0.LastChild = Elements_Object_Value3

			{
			}
			_ = Array_Elements_Value1
			_ = COMMA2
			_ = Elements_Object_Value3
		case 2:
			Elements0 := lhs
			Array_Elements_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Elements_Value3 := rhs[2]

			Elements0.Child = Array_Elements_Value1
			Array_Elements_Value1.Next = COMMA2
			COMMA2.Next = Elements_Value3
			Elements0.LastChild = Elements_Value3

			{
			}
			_ = Array_Elements_Value1
			_ = COMMA2
			_ = Elements_Value3
		case 3:
			Elements0 := lhs
			Elements1 := rhs[0]
			COMMA2 := rhs[1]
			Array_Elements_Value3 := rhs[2]

			Elements0.Child = Elements1
			Elements1.Next = COMMA2
			COMMA2.Next = Array_Elements_Value3
			Elements0.LastChild = Array_Elements_Value3

			{
			}
			_ = Elements1
			_ = COMMA2
			_ = Array_Elements_Value3
		case 4:
			Elements0 := lhs
			Elements1 := rhs[0]
			COMMA2 := rhs[1]
			Elements_Object_Value3 := rhs[2]

			Elements0.Child = Elements1
			Elements1.Next = COMMA2
			COMMA2.Next = Elements_Object_Value3
			Elements0.LastChild = Elements_Object_Value3

			{
			}
			_ = Elements1
			_ = COMMA2
			_ = Elements_Object_Value3
		case 5:
			Elements0 := lhs
			Elements1 := rhs[0]
			COMMA2 := rhs[1]
			Elements_Value3 := rhs[2]

			Elements0.Child = Elements1
			Elements1.Next = COMMA2
			COMMA2.Next = Elements_Value3
			Elements0.LastChild = Elements_Value3

			{
			}
			_ = Elements1
			_ = COMMA2
			_ = Elements_Value3
		case 6:
			Elements0 := lhs
			Elements_Object_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Array_Elements_Value3 := rhs[2]

			Elements0.Child = Elements_Object_Value1
			Elements_Object_Value1.Next = COMMA2
			COMMA2.Next = Array_Elements_Value3
			Elements0.LastChild = Array_Elements_Value3

			{
			}
			_ = Elements_Object_Value1
			_ = COMMA2
			_ = Array_Elements_Value3
		case 7:
			Elements0 := lhs
			Elements_Object_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Elements_Object_Value3 := rhs[2]

			Elements0.Child = Elements_Object_Value1
			Elements_Object_Value1.Next = COMMA2
			COMMA2.Next = Elements_Object_Value3
			Elements0.LastChild = Elements_Object_Value3

			{
			}
			_ = Elements_Object_Value1
			_ = COMMA2
			_ = Elements_Object_Value3
		case 8:
			Elements0 := lhs
			Elements_Object_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Elements_Value3 := rhs[2]

			Elements0.Child = Elements_Object_Value1
			Elements_Object_Value1.Next = COMMA2
			COMMA2.Next = Elements_Value3
			Elements0.LastChild = Elements_Value3

			{
			}
			_ = Elements_Object_Value1
			_ = COMMA2
			_ = Elements_Value3
		case 9:
			Elements0 := lhs
			Elements_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Array_Elements_Value3 := rhs[2]

			Elements0.Child = Elements_Value1
			Elements_Value1.Next = COMMA2
			COMMA2.Next = Array_Elements_Value3
			Elements0.LastChild = Array_Elements_Value3

			{
			}
			_ = Elements_Value1
			_ = COMMA2
			_ = Array_Elements_Value3
		case 10:
			Elements0 := lhs
			Elements_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Elements_Object_Value3 := rhs[2]

			Elements0.Child = Elements_Value1
			Elements_Value1.Next = COMMA2
			COMMA2.Next = Elements_Object_Value3
			Elements0.LastChild = Elements_Object_Value3

			{
			}
			_ = Elements_Value1
			_ = COMMA2
			_ = Elements_Object_Value3
		case 11:
			Elements0 := lhs
			Elements_Value1 := rhs[0]
			COMMA2 := rhs[1]
			Elements_Value3 := rhs[2]

			Elements0.Child = Elements_Value1
			Elements_Value1.Next = COMMA2
			COMMA2.Next = Elements_Value3
			Elements0.LastChild = Elements_Value3

			{
			}
			_ = Elements_Value1
			_ = COMMA2
			_ = Elements_Value3
		case 12:
			Members0 := lhs
			Members1 := rhs[0]
			COMMA2 := rhs[1]
			Members_Pair3 := rhs[2]

			Members0.Child = Members1
			Members1.Next = COMMA2
			COMMA2.Next = Members_Pair3
			Members0.LastChild = Members_Pair3

			{
			}
			_ = Members1
			_ = COMMA2
			_ = Members_Pair3
		case 13:
			Members0 := lhs
			Members_Pair1 := rhs[0]
			COMMA2 := rhs[1]
			Members_Pair3 := rhs[2]

			Members0.Child = Members_Pair1
			Members_Pair1.Next = COMMA2
			COMMA2.Next = Members_Pair3
			Members0.LastChild = Members_Pair3

			{
			}
			_ = Members_Pair1
			_ = COMMA2
			_ = Members_Pair3
		case 14:
			Elements_Value0 := lhs
			BOOL1 := rhs[0]

			Elements_Value0.Child = BOOL1
			Elements_Value0.LastChild = BOOL1

			{
			}
			_ = BOOL1
		case 15:
			Elements_Object_Value0 := lhs
			LCURLY1 := rhs[0]
			Members2 := rhs[1]
			RCURLY3 := rhs[2]

			Elements_Object_Value0.Child = LCURLY1
			LCURLY1.Next = Members2
			Members2.Next = RCURLY3
			Elements_Object_Value0.LastChild = RCURLY3

			{
			}
			_ = LCURLY1
			_ = Members2
			_ = RCURLY3
		case 16:
			Elements_Object_Value0 := lhs
			LCURLY1 := rhs[0]
			Members_Pair2 := rhs[1]
			RCURLY3 := rhs[2]

			Elements_Object_Value0.Child = LCURLY1
			LCURLY1.Next = Members_Pair2
			Members_Pair2.Next = RCURLY3
			Elements_Object_Value0.LastChild = RCURLY3

			{
			}
			_ = LCURLY1
			_ = Members_Pair2
			_ = RCURLY3
		case 17:
			Elements_Object_Value0 := lhs
			LCURLY1 := rhs[0]
			RCURLY2 := rhs[1]

			Elements_Object_Value0.Child = LCURLY1
			LCURLY1.Next = RCURLY2
			Elements_Object_Value0.LastChild = RCURLY2

			{
			}
			_ = LCURLY1
			_ = RCURLY2
		case 18:
			Array_Elements_Value0 := lhs
			LSQUARE1 := rhs[0]
			Array_Elements_Value2 := rhs[1]
			RSQUARE3 := rhs[2]

			Array_Elements_Value0.Child = LSQUARE1
			LSQUARE1.Next = Array_Elements_Value2
			Array_Elements_Value2.Next = RSQUARE3
			Array_Elements_Value0.LastChild = RSQUARE3

			{
			}
			_ = LSQUARE1
			_ = Array_Elements_Value2
			_ = RSQUARE3
		case 19:
			Array_Elements_Value0 := lhs
			LSQUARE1 := rhs[0]
			Elements2 := rhs[1]
			RSQUARE3 := rhs[2]

			Array_Elements_Value0.Child = LSQUARE1
			LSQUARE1.Next = Elements2
			Elements2.Next = RSQUARE3
			Array_Elements_Value0.LastChild = RSQUARE3

			{
			}
			_ = LSQUARE1
			_ = Elements2
			_ = RSQUARE3
		case 20:
			Array_Elements_Value0 := lhs
			LSQUARE1 := rhs[0]
			Elements_Object_Value2 := rhs[1]
			RSQUARE3 := rhs[2]

			Array_Elements_Value0.Child = LSQUARE1
			LSQUARE1.Next = Elements_Object_Value2
			Elements_Object_Value2.Next = RSQUARE3
			Array_Elements_Value0.LastChild = RSQUARE3

			{
			}
			_ = LSQUARE1
			_ = Elements_Object_Value2
			_ = RSQUARE3
		case 21:
			Array_Elements_Value0 := lhs
			LSQUARE1 := rhs[0]
			Elements_Value2 := rhs[1]
			RSQUARE3 := rhs[2]

			Array_Elements_Value0.Child = LSQUARE1
			LSQUARE1.Next = Elements_Value2
			Elements_Value2.Next = RSQUARE3
			Array_Elements_Value0.LastChild = RSQUARE3

			{
			}
			_ = LSQUARE1
			_ = Elements_Value2
			_ = RSQUARE3
		case 22:
			Array_Elements_Value0 := lhs
			LSQUARE1 := rhs[0]
			RSQUARE2 := rhs[1]

			Array_Elements_Value0.Child = LSQUARE1
			LSQUARE1.Next = RSQUARE2
			Array_Elements_Value0.LastChild = RSQUARE2

			{
			}
			_ = LSQUARE1
			_ = RSQUARE2
		case 23:
			Elements_Value0 := lhs
			NULL1 := rhs[0]

			Elements_Value0.Child = NULL1
			Elements_Value0.LastChild = NULL1

			{
			}
			_ = NULL1
		case 24:
			Elements_Value0 := lhs
			NUMBER1 := rhs[0]

			Elements_Value0.Child = NUMBER1
			Elements_Value0.LastChild = NUMBER1

			{
			}
			_ = NUMBER1
		case 25:
			Elements_Value0 := lhs
			STRING1 := rhs[0]

			Elements_Value0.Child = STRING1
			Elements_Value0.LastChild = STRING1

			{
			}
			_ = STRING1
		case 26:
			Members_Pair0 := lhs
			STRING1 := rhs[0]
			COLON2 := rhs[1]
			Array_Elements_Value3 := rhs[2]

			Members_Pair0.Child = STRING1
			STRING1.Next = COLON2
			COLON2.Next = Array_Elements_Value3
			Members_Pair0.LastChild = Array_Elements_Value3

			{
			}
			_ = STRING1
			_ = COLON2
			_ = Array_Elements_Value3
		case 27:
			Members_Pair0 := lhs
			STRING1 := rhs[0]
			COLON2 := rhs[1]
			Elements_Object_Value3 := rhs[2]

			Members_Pair0.Child = STRING1
			STRING1.Next = COLON2
			COLON2.Next = Elements_Object_Value3
			Members_Pair0.LastChild = Elements_Object_Value3

			{
			}
			_ = STRING1
			_ = COLON2
			_ = Elements_Object_Value3
		case 28:
			Members_Pair0 := lhs
			STRING1 := rhs[0]
			COLON2 := rhs[1]
			Elements_Value3 := rhs[2]

			Members_Pair0.Child = STRING1
			STRING1.Next = COLON2
			COLON2.Next = Elements_Value3
			Members_Pair0.LastChild = Elements_Value3

			{
			}
			_ = STRING1
			_ = COLON2
			_ = Elements_Value3
		}
		_ = ruleFlags
	}

	return &gopapageno.Grammar{
		NumTerminals:              numTerminals,
		NumNonterminals:           numNonTerminals,
		MaxRHSLength:              maxRHSLen,
		Rules:                     rules,
		CompressedRules:           compressedRules,
		PrecedenceMatrix:          precMatrix,
		BitPackedPrecedenceMatrix: bitPackedMatrix,
		Func:                      fn,
		ParsingStrategy:           gopapageno.OPP,
	}
}
