package gopapageno

import (
	"fmt"
	"strings"
)

type Token struct {
	Type       TokenType
	Precedence Precedence

	Value any
	// Lexeme string

	Next  *Token
	Child *Token
}

func (t *Token) IsTerminal() bool {
	return t.Type.IsTerminal()
}

type TokenType uint16

const (
	TokenEmpty TokenType = 0
	TokenTerm  TokenType = 0x8000
)

func (t TokenType) IsTerminal() bool {
	return t >= 0x8000
}

func (t TokenType) Value() uint16 {
	return uint16(0x7FFF & t)
}

// Height computes the height of the AST rooted in `t`.
// It can be used as an evaluation metric for tree-balance, as left/right-skewed trees will have a bigger height compared to balanced trees.
func (t *Token) Height() int {
	var rec func(t *Token, root bool) int

	rec = func(t *Token, root bool) int {
		if t == nil {
			return 0
		}

		if root {
			return 1 + rec(t.Child, false)
		} else {
			return max(1+rec(t.Child, false), rec(t.Next, false))
		}
	}

	return rec(t, true)
}

// Size returns the number of tokens in the AST rooted in `t`.
func (t *Token) Size() int {
	var rec func(t *Token, root bool) int

	rec = func(t *Token, root bool) int {
		if t == nil {
			return 0
		}

		if root {
			return 1 + rec(t.Child, false)
		} else {
			return 1 + rec(t.Child, false) + rec(t.Next, false)
		}
	}

	return rec(t, true)
}

// String returns a string representation of the AST rooted in `t`.
// This should be used rarely, as it doesn't print out a proper string representation of the token type.
// Gopapageno will generate a `SprintToken` function for your tokens.
func (t *Token) String() string {
	var sprintRec func(t *Token, sb *strings.Builder, indent string)

	sprintRec = func(t *Token, sb *strings.Builder, indent string) {
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

		sb.WriteString(fmt.Sprintf("%d: %v\n", t.Type, t.Value))

		sprintRec(t.Child, sb, indent)
		sprintRec(t.Next, sb, indent[:len(indent)-4])
	}

	var sb strings.Builder

	sprintRec(t, &sb, "")

	return sb.String()
}
