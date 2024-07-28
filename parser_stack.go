package gopapageno

import (
	"fmt"
)

// parserStack is the base data structure used by workers during parsing.
type parserStack struct {
	*LOPS[Token]

	firstTerminal *Token
}

// newParserStack creates an empty parserStack.
func newParserStack(pool *Pool[stack[*Token]]) *parserStack {
	return &parserStack{
		LOPS: NewLOPS[Token](pool),
	}
}

// Push pushes a token pointer in the ParserStack.
// It returns the pointer itself.
func (s *parserStack) Push(token *Token) *Token {
	s.LOPS.Push(token)

	//If the token is a terminal update the firstTerminal pointer
	if token.Type.IsTerminal() {
		s.firstTerminal = token
	}

	return token
}

// FirstTerminal returns a pointer to the first terminal token on the stack.
func (s *parserStack) FirstTerminal() *Token {
	return s.firstTerminal
}

// UpdateFirstTerminal should be used after a reduction in order to update the first terminal counter.
// In fact, in order to save some time, only the Push operation automatically updates the first terminal pointer,
// while the Pop operation does not.
func (s *parserStack) UpdateFirstTerminal() {
	s.firstTerminal = s.findFirstTerminal()
}

// findFirstTerminal computes the first terminal on the stacks.
// This function is for internal usage only.
func (s *parserStack) findFirstTerminal() *Token {
	curStack := s.cur

	pos := curStack.Tos - 1

	for pos < 0 {
		pos = -1
		if curStack.Prev == nil {
			return nil
		}
		curStack = curStack.Prev
		pos = curStack.Tos - 1
	}

	for !curStack.Data[pos].Type.IsTerminal() {
		pos--
		for pos < 0 {
			pos = -1
			if curStack.Prev == nil {
				return nil
			}
			curStack = curStack.Prev
			pos = curStack.Tos - 1
		}
	}

	return curStack.Data[pos]
}

func (s *parserStack) LastNonterminal() (*Token, error) {
	for token := s.Pop(); token != nil; token = s.Pop() {
		if !token.Type.IsTerminal() {
			return token, nil
		}
	}

	return nil, fmt.Errorf("no nonterminal found")
}
