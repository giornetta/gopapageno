package gopapageno

import (
	"fmt"
)

type COPPStack struct {
	*ParserStackBase[CToken]

	StateTokenStack *stack[*CToken]
	State           *CyclicAutomataState
}

// NewCOPPStack creates a new COPPStack initialized with one empty stack.
func NewCOPPStack(tokenStackPool *Pool[stack[*CToken]]) *COPPStack {
	return &COPPStack{
		ParserStackBase: NewParserStack(tokenStackPool),
		StateTokenStack: tokenStackPool.Get(),

		State: new(CyclicAutomataState),
	}
}

func (s *COPPStack) Current() []*CToken {
	return s.StateTokenStack.Slice(s.State.CurrentIndex, s.State.CurrentLen)
}

func (s *COPPStack) Previous() []*CToken {
	return s.StateTokenStack.Slice(s.State.PreviousIndex, s.State.PreviousLen)
}

func (s *COPPStack) IsCurrentSingleNonterminal() bool {
	return s.State.CurrentLen == 1 && !s.StateTokenStack.Data[s.State.CurrentIndex].IsTerminal()
}

func (s *COPPStack) AppendStateToken(token *CToken) {
	s.StateTokenStack.Push(token)
	s.State.CurrentLen++
}

func (s *COPPStack) SwapState() {
	s.State.PreviousIndex, s.State.PreviousLen = s.State.CurrentIndex, s.State.CurrentLen

	s.State.CurrentIndex = s.StateTokenStack.Tos
	s.State.CurrentLen = 0
}

func (s *COPPStack) Push(token *CToken) *CToken {
	token.State = *s.State

	t := s.ParserStackBase.Push(token)

	return t
}

func (s *COPPStack) PushWithState(token *CToken) *CToken {
	t := s.ParserStackBase.Push(token)

	return t
}

func (s *COPPStack) IsYieldingPrecedence() bool {
	return s.firstTerminal.Precedence == PrecYields || s.firstTerminal.Precedence == PrecEquals
}

func (s *COPPStack) Pop() *CToken {
	token := s.ParserStackBase.Pop()

	return token
}

func (s *COPPStack) Combine() *COPPStack {
	var tlStack *stack[*CToken]

	var tlPosition int
	removedTokens := -1

	// TODO: This could be moved in Push/Pop to allow constant time access.
	it := s.HeadIterator()
	first := true
	for t := it.Next(); t != nil && ((t.Precedence != PrecYields && t.Precedence != PrecEquals) || (first && t.Type != TokenTerm)); t = it.Next() {
		first = false

		tlStack = it.cur
		tlPosition = it.pos

		removedTokens++
	}

	if s.cur.Data[tlPosition].Type != TokenEmpty {
		s.cur.Data[tlPosition].Precedence = PrecEmpty
	}

	s.ParserStackBase.head = tlStack
	s.ParserStackBase.headFirst = tlPosition
	s.ParserStackBase.len -= removedTokens

	for t := it.Cur(); t != nil && t.Precedence != PrecTakes; t = it.Next() {
		tlPosition = it.pos
	}

	s.ParserStackBase.cur.Tos = tlPosition + 1

	s.UpdateFirstTerminal()

	return s
}

func (s *COPPStack) CombineLOS(pool *Pool[stack[CToken]]) *LOS[CToken] {
	list := NewLOS(pool)

	it := s.HeadIterator()
	t := it.Next()

	tokenSet := make(map[*CToken]struct{}, s.Length())
	tokenSet[t] = struct{}{}
	for _, t := range s.StateTokenStack.Slice(t.State.CurrentIndex, t.State.CurrentLen) {
		tokenSet[t] = struct{}{}
	}

	if s.Length() == 1 {
		for _, t := range s.StateTokenStack.Slice(s.State.CurrentIndex, s.State.CurrentLen) {
			t.Precedence = PrecEmpty
			list.Push(*t)
		}

		return list
	}

	for t := it.Next(); t != nil && (t.Precedence != PrecYields && t.Precedence != PrecEquals); t = it.Next() {
		for _, stateToken := range s.StateTokenStack.Slice(t.State.CurrentIndex, t.State.CurrentLen) {
			_, ok := tokenSet[stateToken]
			if !ok {
				stateToken.Precedence = PrecEmpty
				list.Push(*stateToken)

				tokenSet[stateToken] = struct{}{}
			}
		}

		_, ok := tokenSet[t]
		if !ok {
			t.Precedence = PrecEmpty
			list.Push(*t)

			tokenSet[t] = struct{}{}
		}
	}

	return list
}

func (s *COPPStack) LastNonterminal() (*CToken, error) {
	if s.State.CurrentLen >= 1 {
		return s.StateTokenStack.Slice(s.State.CurrentIndex, s.State.CurrentLen)[0], nil
	}

	return nil, fmt.Errorf("no token stack current")
}
