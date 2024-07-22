package gopapageno

import "fmt"

type OPPStack struct {
	*ParserStackBase[Token]

	yieldingPrec int
}

func NewOPPStack(pool *Pool[stack[*Token]]) *OPPStack {
	return &OPPStack{
		ParserStackBase: NewParserStack(pool),
	}
}

func (s *OPPStack) Push(token *Token) *Token {
	t := s.ParserStackBase.Push(token)

	// If the token is yielding precedence, increase the counter
	if token.Precedence == PrecYields || token.Precedence == PrecAssociative {
		s.yieldingPrec++
	}

	return t
}

func (s *OPPStack) Pop() *Token {
	t := s.ParserStackBase.Pop()

	if t.Precedence == PrecYields || t.Precedence == PrecAssociative {
		s.yieldingPrec--
	}

	return t
}

func (s *OPPStack) YieldingPrecedence() int {
	return s.yieldingPrec
}

func (s *OPPStack) Combine() *OPPStack {
	var topLeft Token

	// TODO: This could be moved in Push/Pop to allow constant time access.
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
