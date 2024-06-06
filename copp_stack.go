package gopapageno

import (
	"fmt"
)

type CyclicAutomataState struct {
	Current  []*Token
	Previous []*Token
}

func NewCyclicAutomataState(maxLength int) *CyclicAutomataState {
	return &CyclicAutomataState{
		Current:  make([]*Token, 0, maxLength),
		Previous: make([]*Token, 0, maxLength),
	}
}

func NewCyclicAutomataStateBuilder(maxLength int) func() *CyclicAutomataState {
	return func() *CyclicAutomataState {
		return &CyclicAutomataState{
			Current:  make([]*Token, 0, maxLength),
			Previous: make([]*Token, 0, maxLength),
		}
	}
}

func NewCyclicAutomataStateValueBuilder(maxLength int) func() CyclicAutomataState {
	return func() CyclicAutomataState {
		return CyclicAutomataState{
			Current:  make([]*Token, 0, maxLength),
			Previous: make([]*Token, 0, maxLength),
		}
	}
}

type CyclicParserStack struct {
	*ParserStack

	StatesLOS *ListOfStacks[CyclicAutomataState]
	State     *CyclicAutomataState

	maxRhsLen int
}

// NewCyclicParserStack creates a new CyclicParserStack initialized with one empty stack.
func NewCyclicParserStack(tokenStackPool *Pool[stack[*Token]], stateStackPool *Pool[stack[CyclicAutomataState]], maxRhsLen int) *CyclicParserStack {
	return &CyclicParserStack{
		ParserStack: NewParserStack(tokenStackPool),
		StatesLOS:   NewListOfStacks[CyclicAutomataState](stateStackPool),

		State:     NewCyclicAutomataState(maxRhsLen),
		maxRhsLen: maxRhsLen,
	}
}

func (s *CyclicParserStack) Push(token *Token, state CyclicAutomataState) *Token {
	t := s.ParserStack.Push(token)

	// To avoid allocations, we load the next available state in the LOS
	// Clear it and append the right elements.
	st := s.StatesLOS.GetNext()

	st.Current = st.Current[:0]
	st.Current = append(st.Current, state.Current...)

	st.Previous = st.Previous[:0]
	st.Previous = append(st.Previous, state.Previous...)

	s.StatesLOS.Push(*st)

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

func (s *CyclicParserStack) Combine(o Stacker) Stacker {
	var topLeft Token
	var topLeftState CyclicAutomataState

	// TODO: This could be moved in Push/Pop to allow constant time access.
	it := s.Iterator()
	first := true
	for t, s := it.Next(); t != nil && ((t.Precedence != PrecYields && t.Precedence != PrecEquals) || (first && t.Type != TokenTerm)); t, s = it.Next() {
		topLeft = *t
		topLeftState = *s

		first = false
	}

	stack := NewCyclicParserStack(s.ParserStack.pool, s.StatesLOS.pool, s.maxRhsLen)

	if topLeft.Type != TokenEmpty {
		topLeft.Precedence = PrecEmpty
		stack.Push(&topLeft, topLeftState)
	}

	for t, s := it.Cur(); t != nil && t.Precedence != PrecTakes; t, s = it.Next() {
		stack.Push(t, *s)
	}

	stack.UpdateFirstTerminal()

	stack.State.Previous = append(stack.State.Previous, s.State.Previous...)
	stack.State.Current = append(stack.State.Current, s.State.Current...)

	return stack
}

func (s *CyclicParserStack) CombineLOS(pool *Pool[stack[Token]]) *ListOfStacks[Token] {
	list := NewListOfStacks[Token](pool)

	it := s.Iterator()
	t, st := it.Next()

	tokenSet := make(map[*Token]struct{}, s.Length())
	tokenSet[t] = struct{}{}
	for _, t := range st.Current {
		tokenSet[t] = struct{}{}
	}

	if s.Length() == 1 {
		for _, t := range s.State.Current {
			t.Precedence = PrecEmpty
			list.Push(*t)
		}

		return list
	}

	for t, st := it.Next(); t != nil && (t.Precedence != PrecYields && t.Precedence != PrecEquals); t, st = it.Next() {
		for _, stateToken := range st.Current {
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
	if len(s.State.Current) >= 1 {
		return s.State.Current[0], nil
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
