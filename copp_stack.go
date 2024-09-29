package gopapageno

import (
	"fmt"
)

type CyclicAutomataState struct {
	CurrentIndex int
	CurrentLen   int

	PreviousIndex int
	PreviousLen   int
}

type COPPStack struct {
	*parserStack

	StatesLOS       *LOS[CyclicAutomataState]
	StateTokenStack *stack[*Token]

	State CyclicAutomataState

	ProducedTokens map[*Token]*Token

	yieldsPrec int
}

// NewCOPPStack creates a new COPPStack initialized with one empty stack.
func NewCOPPStack(tokenStackPool *Pool[stack[*Token]], stateStackPool *Pool[stack[CyclicAutomataState]], producedTokens map[*Token]*Token) *COPPStack {
	return &COPPStack{
		parserStack:     newParserStack(tokenStackPool),
		StatesLOS:       NewLOS[CyclicAutomataState](stateStackPool),
		StateTokenStack: tokenStackPool.Get(),

		// ProducedTokens maps the leftmost rhs-token of a prefix production with its temporary parent.
		ProducedTokens: producedTokens,

		State: CyclicAutomataState{},
	}
}

func (s *COPPStack) Current() []*Token {
	return s.StateTokenStack.Slice(s.State.CurrentIndex, s.State.CurrentLen)
}

func (s *COPPStack) Previous() []*Token {
	return s.StateTokenStack.Slice(s.State.PreviousIndex, s.State.PreviousLen)
}

func (s *COPPStack) IsCurrentSingleNonterminal() bool {
	return s.State.CurrentLen == 1 && !s.StateTokenStack.Data[s.State.CurrentIndex].IsTerminal()
}

func (s *COPPStack) AppendStateToken(token *Token) {
	s.StateTokenStack.Push(token)
	s.State.CurrentLen++
}

func (s *COPPStack) SwapState() {
	s.State.PreviousIndex, s.State.PreviousLen = s.State.CurrentIndex, s.State.CurrentLen

	s.State.CurrentIndex = s.StateTokenStack.Tos
	s.State.CurrentLen = 0
}

func (s *COPPStack) Push(token *Token) *Token {
	t := s.parserStack.Push(token)
	s.StatesLOS.Push(s.State)

	return t
}

func (s *COPPStack) PushWithState(token *Token, state CyclicAutomataState) *Token {
	t := s.parserStack.Push(token)
	s.StatesLOS.Push(state)

	return t
}

func (s *COPPStack) YieldingPrecedence() int {
	ft := s.FirstTerminal()
	if ft.Precedence != PrecYields && ft.Precedence != PrecEquals {
		return 0
	}

	if s.firstTerminalPos <= s.headFirst {
		return 1
	}

	lt := s.firstTerminalStack.Data[s.firstTerminalPos-1]
	if (lt.Precedence == PrecTakes || lt.Precedence == PrecEmpty) && lt.Type == ft.Type {
		return 0
	}

	return 1
}

func (s *COPPStack) Pop2() (*Token, *CyclicAutomataState) {
	token := s.parserStack.Pop()
	state := s.StatesLOS.Pop()

	return token, state
}

func (s *COPPStack) Pop() *Token {
	token := s.parserStack.Pop()
	_ = s.StatesLOS.Pop()

	return token
}

func (s *COPPStack) Combine() *COPPStack {
	var tlStack *stack[*Token]
	var tlStStack *stack[CyclicAutomataState]

	var tlPosition int
	removedTokens := -1

	it := s.Iterator()
	first := true
	for t, _ := it.Next(); t != nil && ((t.Precedence != PrecYields && t.Precedence != PrecEquals) || first); t, _ = it.Next() {
		first = false

		tlStack = it.TokensIt.cur
		tlStStack = it.StatesIt.cur

		tlPosition = it.TokensIt.pos

		removedTokens++
	}

	if s.cur.Data[tlPosition].Type != TokenEmpty {
		s.cur.Data[tlPosition].Precedence = PrecEmpty
	}

	s.parserStack.head = tlStack
	s.StatesLOS.head = tlStStack

	s.parserStack.headFirst = tlPosition
	s.StatesLOS.headFirst = tlPosition

	s.parserStack.head.Prev = nil
	s.StatesLOS.head.Prev = nil

	s.parserStack.len -= removedTokens
	s.StatesLOS.len -= removedTokens

	for t, _ := it.Cur(); t != nil && t.Precedence != PrecTakes; t, _ = it.Next() {
		tlPosition = it.TokensIt.pos
	}

	s.parserStack.cur.Tos = tlPosition + 1
	s.StatesLOS.cur.Tos = tlPosition + 1

	s.UpdateFirstTerminal()

	return s
}

func (s *COPPStack) CombineLOS(pool *Pool[stack[Token]]) (*LOS[Token], map[*Token]*Token) {
	list := NewLOS[Token](pool)
	newProducedTokens := make(map[*Token]*Token)

	it := s.Iterator()

	t, st := it.Next()
	tokenSet := make(map[*Token]struct{}, s.Length())
	tokenSet[t] = struct{}{}
	for _, t := range s.StateTokenStack.Slice(st.CurrentIndex, st.CurrentLen) {
		tokenSet[t] = struct{}{}
	}

	for t, st = it.Next(); t != nil && t.Precedence != PrecYields; t, st = it.Next() {
		if t.Precedence == PrecEquals {
			if !it.IsLast() {
				continue
			}

			for _, stateToken := range s.StateTokenStack.Slice(st.CurrentIndex, st.CurrentLen) {
				if _, ok := tokenSet[stateToken]; !ok {
					stateToken.Precedence = PrecEmpty

					newToken := list.Push(*stateToken)
					parentToken, ok := s.ProducedTokens[stateToken]
					if ok {
						newProducedTokens[newToken] = parentToken
					}

					tokenSet[stateToken] = struct{}{}
				}
			}

			for _, stateToken := range s.Previous() {
				if _, ok := tokenSet[stateToken]; !ok {
					stateToken.Precedence = PrecEmpty

					newToken := list.Push(*stateToken)
					parentToken, ok := s.ProducedTokens[stateToken]
					if ok {
						newProducedTokens[newToken] = parentToken
					}

					tokenSet[stateToken] = struct{}{}
				}
			}

			for _, stateToken := range s.Current() {
				if _, ok := tokenSet[stateToken]; !ok {
					stateToken.Precedence = PrecEmpty

					newToken := list.Push(*stateToken)
					parentToken, ok := s.ProducedTokens[stateToken]
					if ok {
						newProducedTokens[newToken] = parentToken
					}

					tokenSet[stateToken] = struct{}{}
				}
			}

			continue
		}

		for _, stateToken := range s.StateTokenStack.Slice(st.CurrentIndex, st.CurrentLen) {
			if _, ok := tokenSet[stateToken]; !ok {
				stateToken.Precedence = PrecEmpty

				newToken := list.Push(*stateToken)
				parentToken, ok := s.ProducedTokens[stateToken]
				if ok {
					newProducedTokens[newToken] = parentToken
				}

				tokenSet[stateToken] = struct{}{}
			}
		}

		if _, ok := tokenSet[t]; !ok {
			t.Precedence = PrecEmpty

			newToken := list.Push(*t)
			parentToken, ok := s.ProducedTokens[t]
			if ok {
				newProducedTokens[newToken] = parentToken
			}

			tokenSet[t] = struct{}{}
		}
	}

	return list, newProducedTokens
}

func (s *COPPStack) LastNonterminal() (*Token, error) {
	if s.State.CurrentLen >= 1 {
		return s.StateTokenStack.Slice(s.State.CurrentIndex, s.State.CurrentLen)[0], nil
	}

	return nil, fmt.Errorf("no token stack current")
}

func (s *COPPStack) Iterator() *CyclicParserStackIterator {
	return &CyclicParserStackIterator{
		TokensIt: s.parserStack.HeadIterator(),
		StatesIt: s.StatesLOS.HeadIterator(),
	}
}

type CyclicParserStackIterator struct {
	TokensIt *LOPSIt[Token]
	StatesIt *LOSIt[CyclicAutomataState]
}

func (i *CyclicParserStackIterator) Next() (*Token, *CyclicAutomataState) {
	return i.TokensIt.Next(), i.StatesIt.Next()
}

func (i *CyclicParserStackIterator) Cur() (*Token, *CyclicAutomataState) {
	return i.TokensIt.Cur(), i.StatesIt.Cur()
}

func (i *CyclicParserStackIterator) IsLast() bool {
	return i.TokensIt.IsLast()
}
