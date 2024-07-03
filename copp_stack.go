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

type CyclicParserStack struct {
	*ParserStack

	StatesLOS       *ListOfStacks[CyclicAutomataState]
	StateTokenStack *stack[*Token]
	State           *CyclicAutomataState
}

// NewCyclicParserStack creates a new CyclicParserStack initialized with one empty stack.
func NewCyclicParserStack(tokenStackPool *Pool[stack[*Token]], stateStackPool *Pool[stack[CyclicAutomataState]]) *CyclicParserStack {
	return &CyclicParserStack{
		ParserStack:     NewParserStack(tokenStackPool),
		StatesLOS:       NewListOfStacks[CyclicAutomataState](stateStackPool),
		StateTokenStack: tokenStackPool.Get(),

		State: new(CyclicAutomataState),
	}
}

func (s *CyclicParserStack) Current() []*Token {
	return s.StateTokenStack.Slice(s.State.CurrentIndex, s.State.CurrentLen)
}

func (s *CyclicParserStack) Previous() []*Token {
	return s.StateTokenStack.Slice(s.State.PreviousIndex, s.State.PreviousLen)
}

func (s *CyclicParserStack) IsCurrentSingleNonterminal() bool {
	return s.State.CurrentLen == 1 && !s.StateTokenStack.Data[s.State.CurrentIndex].IsTerminal()
}

func (s *CyclicParserStack) AppendStateToken(token *Token) {
	s.StateTokenStack.Push(token)
	s.State.CurrentLen++
}

func (s *CyclicParserStack) SwapState() {
	s.State.PreviousIndex, s.State.PreviousLen = s.State.CurrentIndex, s.State.CurrentLen

	s.State.CurrentIndex = s.StateTokenStack.Tos
	s.State.CurrentLen = 0
}

func (s *CyclicParserStack) Push(token *Token) *Token {
	t := s.ParserStack.Push(token)
	s.StatesLOS.Push(*s.State)

	return t
}

func (s *CyclicParserStack) PushWithState(token *Token, state CyclicAutomataState) *Token {
	t := s.ParserStack.Push(token)
	s.StatesLOS.Push(state)

	return t
}

func (s *CyclicParserStack) YieldingPrecedence() int {
	if s.firstTerminal.Precedence == PrecYields || s.firstTerminal.Precedence == PrecEquals {
		return 1
	}

	return 0
}

func (s *CyclicParserStack) Pop2() (*Token, *CyclicAutomataState) {
	token := s.ParserStack.Pop()
	state := s.StatesLOS.Pop()

	return token, state
}

func (s *CyclicParserStack) Pop() *Token {
	token := s.ParserStack.Pop()
	_ = s.StatesLOS.Pop()

	return token
}

// Merge links the stacks of the current and of another ParserStack.
func (s *CyclicParserStack) Merge(other *CyclicParserStack) {
	s.ParserStack.Merge(other.ParserStack)
	s.StatesLOS.Merge(other.StatesLOS)
}

func (s *CyclicParserStack) Split(n int) ([]*CyclicParserStack, error) {
	stacks, err := s.ParserStack.Split(n)
	if err != nil {
		return nil, fmt.Errorf("could not split token stack: %w", err)
	}

	states, err := s.StatesLOS.Split(n)
	if err != nil {
		return nil, fmt.Errorf("could not split states stack: %w", err)
	}

	newStacks := make([]*CyclicParserStack, len(stacks))
	for i, _ := range stacks {
		newStacks[i] = &CyclicParserStack{
			ParserStack: stacks[i],
			StatesLOS:   states[i],
		}
	}

	return newStacks, nil
}

func (s *CyclicParserStack) Combine() Stacker {
	var tlStack *stack[*Token]
	var tlStStack *stack[CyclicAutomataState]

	var tlPosition int
	removedTokens := -1

	// TODO: This could be moved in Push/Pop to allow constant time access.
	it := s.Iterator()
	first := true
	for t, _ := it.Next(); t != nil && ((t.Precedence != PrecYields && t.Precedence != PrecEquals) || (first && t.Type != TokenTerm)); t, _ = it.Next() {
		first = false

		tlStack = it.TokensIt.cur
		tlStStack = it.StatesIt.cur

		tlPosition = it.TokensIt.pos

		removedTokens++
	}

	if s.cur.Data[tlPosition].Type != TokenEmpty {
		s.cur.Data[tlPosition].Precedence = PrecEmpty
	}

	s.ParserStack.head = tlStack
	s.StatesLOS.head = tlStStack

	s.ParserStack.headFirst = tlPosition
	s.StatesLOS.headFirst = tlPosition

	s.ParserStack.len -= removedTokens
	s.StatesLOS.len -= removedTokens

	for t, _ := it.Cur(); t != nil && t.Precedence != PrecTakes; t, _ = it.Next() {
		tlPosition = it.TokensIt.pos
	}

	s.ParserStack.cur.Tos = tlPosition + 1
	s.StatesLOS.cur.Tos = tlPosition + 1

	s.UpdateFirstTerminal()

	// stack.State = s.State

	return s
}

func (s *CyclicParserStack) CombineLOS(pool *Pool[stack[Token]]) *ListOfStacks[Token] {
	list := NewListOfStacks[Token](pool)

	it := s.Iterator()
	t, st := it.Next()

	tokenSet := make(map[*Token]struct{}, s.Length())
	tokenSet[t] = struct{}{}
	for _, t := range s.StateTokenStack.Slice(st.CurrentIndex, st.CurrentLen) {
		tokenSet[t] = struct{}{}
	}

	if s.Length() == 1 {
		for _, t := range s.StateTokenStack.Slice(s.State.CurrentIndex, s.State.CurrentLen) {
			t.Precedence = PrecEmpty
			list.Push(*t)
		}

		return list
	}

	for t, st := it.Next(); t != nil && (t.Precedence != PrecYields && t.Precedence != PrecEquals); t, st = it.Next() {
		for _, stateToken := range s.StateTokenStack.Slice(st.CurrentIndex, st.CurrentLen) {
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

func (s *CyclicParserStack) LastNonterminal() (*Token, error) {
	if s.State.CurrentLen >= 1 {
		return s.StateTokenStack.Slice(s.State.CurrentIndex, s.State.CurrentLen)[0], nil
	}

	return nil, fmt.Errorf("no token stack current")
}

func (s *CyclicParserStack) Iterator() *CyclicParserStackIterator {
	return &CyclicParserStackIterator{
		TokensIt: s.ParserStack.HeadIterator(),
		StatesIt: s.StatesLOS.HeadIterator(),
	}
}

type CyclicParserStackIterator struct {
	TokensIt *ParserStackIterator
	StatesIt *LosIterator[CyclicAutomataState]
}

func (i *CyclicParserStackIterator) Next() (*Token, *CyclicAutomataState) {
	return i.TokensIt.Next(), i.StatesIt.Next()
}

func (i *CyclicParserStackIterator) Cur() (*Token, *CyclicAutomataState) {
	return i.TokensIt.Cur(), i.StatesIt.Cur()
}

func (i *CyclicParserStackIterator) IsLast() bool {
	if i.TokensIt.pos+1 < i.TokensIt.cur.Tos {
		return false
	}

	if i.TokensIt.cur.Next == nil {
		return true
	}

	return false
}
