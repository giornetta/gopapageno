package gopapageno

import (
	"context"
	"fmt"
	"strings"
)

type Tokener interface {
	IsTerminal() bool
	GetValue() any
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

type TokenBase[T Tokener] struct {
	Type       TokenType
	Precedence Precedence

	Value any

	Next  *T
	Child *T
}

func (t TokenBase[T]) IsTerminal() bool {
	return t.Type.IsTerminal()
}

func (t TokenBase[T]) GetValue() any {
	return t.Value
}

type Token struct {
	TokenBase[Token]
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

// Height computes the height of the AST rooted in `t`.
// It can be used as an evaluation metric for tree-balance, as left/right-skewed trees will have a bigger height compared to balanced trees.
func (t *Token) Height(ctx context.Context) (int, error) {
	return 1, nil
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

type CToken struct {
	TokenBase[CToken]

	LastChild *CToken
	State     CyclicAutomataState
}

type CyclicAutomataState struct {
	CurrentIndex int
	CurrentLen   int

	PreviousIndex int
	PreviousLen   int
}

// Size returns the number of tokens in the AST rooted in `t`.
func (t *CToken) Size() int {
	var rec func(t *CToken, root bool) int

	rec = func(t *CToken, root bool) int {
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

// Height computes the height of the AST rooted in `t`.
// It can be used as an evaluation metric for tree-balance, as left/right-skewed trees will have a bigger height compared to balanced trees.
func (t *CToken) Height(ctx context.Context) (int, error) {
	return 1, nil
}

// String returns a string representation of the AST rooted in `t`.
// This should be used rarely, as it doesn't print out a proper string representation of the token type.
// Gopapageno will generate a `SprintToken` function for your tokens.
func (t *CToken) String() string {
	var sprintRec func(t *CToken, sb *strings.Builder, indent string)

	sprintRec = func(t *CToken, sb *strings.Builder, indent string) {
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
