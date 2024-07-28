package gopapageno

import (
	"fmt"
)

// OPPStack is the data structure used by OPP and AOPP workers during parsing.
type OPPStack struct {
	*parserStack

	yieldsPrec int
}

func NewOPPStack(pool *Pool[stack[*Token]]) *OPPStack {
	return &OPPStack{
		parserStack: newParserStack(pool),
	}
}

// FirstTerminal returns a pointer to the first terminal token on the stack.
func (s *OPPStack) FirstTerminal() *Token {
	return s.firstTerminal
}

// UpdateFirstTerminal should be used after a reduction in order to update the first terminal counter.
// In fact, in order to save some time, only the Push operation automatically updates the first terminal pointer,
// while the Pop operation does not.
func (s *OPPStack) UpdateFirstTerminal() {
	s.firstTerminal = s.findFirstTerminal()
}

// findFirstTerminal computes the first terminal on the stacks.
// This function is for internal usage only.
func (s *OPPStack) findFirstTerminal() *Token {
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

// Push pushes a token pointer in the parserStack.
// It returns the pointer itself.
func (s *OPPStack) Push(token *Token) *Token {
	s.parserStack.Push(token)

	// If the token is yielding precedence, increase the counter
	if token.Precedence == PrecYields || token.Precedence == PrecAssociative {
		s.yieldsPrec++
	}

	return token
}

// Pop removes the topmost element from the parserStack and returns it.
func (s *OPPStack) Pop() *Token {
	t := s.parserStack.Pop()
	if t == nil {
		return nil
	}

	if t.Precedence == PrecYields || t.Precedence == PrecAssociative {
		s.yieldsPrec--
	}

	return t
}

func (s *OPPStack) YieldingPrecedence() int {
	return s.yieldsPrec
}

func (s *OPPStack) Combine() *OPPStack {
	var topLeft Token

	it := s.HeadIterator()
	for t := it.Next(); t != nil && t.Precedence != PrecYields; t = it.Next() {
		topLeft = *t
	}

	list := NewOPPStack(s.pool)

	topLeft.Precedence = PrecEmpty
	list.Push(&topLeft)

	for t := it.Cur(); t != nil && t.Precedence != PrecTakes; t = it.Next() {
		list.Push(t)
	}

	list.UpdateFirstTerminal()

	return list
}

func (s *OPPStack) CombineLOS(pool *Pool[stack[Token]]) *LOS[Token] {
	var tok Token

	list := NewLOS[Token](pool)

	it := s.HeadIterator()

	// Ignore first element
	it.Next()

	for t := it.Next(); t != nil && t.Precedence != PrecYields; t = it.Next() {
		tok = *t
		tok.Precedence = PrecEmpty
		list.Push(tok)
	}

	return list
}

func (s *OPPStack) LastNonterminal() (*Token, error) {
	for token := s.Pop(); token != nil; token = s.Pop() {
		if !token.Type.IsTerminal() {
			return token, nil
		}
	}

	return nil, fmt.Errorf("no nonterminal found")
}
