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
		State:       NewCyclicAutomataState(maxRhsLen),
		maxRhsLen:   maxRhsLen,
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
	if s.firstTerminal.Precedence == PrecYields {
		return s.yieldsPrec
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
	for t, s := it.Next(); t != nil && t.Precedence != PrecYields; t, s = it.Next() {
		topLeft = *t
		topLeftState = *s
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

	os := o.(*CyclicParserStack)
	oit := os.Iterator()

	tok, state := oit.Next()
	if tok.Type == TokenTerm {
		stack.State.Previous = append([]*Token{}, s.State.Current...)
		stack.State.Current = append([]*Token{}, state.Current...)
	} else {
		var lastState *CyclicAutomataState
		tok, state = oit.Next()
		if tok != nil && tok.Precedence != PrecYields {
			lastState = state
		}

		// Other stack only has the first "forced" token in it.
		// It means it managed to reduce everything about its input chunk.
		if lastState == nil || len(lastState.Current) == 0 {
			stack.State.Previous = append(stack.State.Previous, s.State.Previous...)
			stack.State.Current = append(stack.State.Current, s.State.Current...)
		} else {
			stack.State.Previous = append(stack.State.Previous, s.State.Current...)
			stack.State.Current = append(stack.State.Current, lastState.Current...)
		}
	}

	return stack
}

func (s *CyclicParserStack) CombineLOS(l *ListOfStacks[Token]) *ListOfStacks[Token] {
	list := NewListOfStacks[Token](l.pool)

	it := s.Iterator()

	// Ignore first element
	t, _ := it.Next()
	if t.Type == TokenTerm {
		t.Precedence = PrecEmpty
		list.Push(*t)

		return list
	}

	var listToken *Token

	first := true
	for t, st := it.Next(); t != nil && t.Precedence != PrecYields; t, st = it.Next() {
		if t.Type == TokenTerm && !first {
			listIt := list.TailIterator()
			for i := len(st.Current) - 1; i >= 0; i-- {
				listToken = listIt.Prev()
				if listToken != nil && listToken == st.Current[i] {
					list.Pop()
				}
			}

			for _, t := range st.Current {
				t.Precedence = PrecEmpty
				list.Push(*t)
			}
		}

		t.Precedence = PrecEmpty
		list.Push(*t)

		first = false
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
