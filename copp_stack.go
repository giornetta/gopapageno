package gopapageno

import "fmt"

type CyclicAutomataState struct {
	Current    []*Token
	CurrentLen int

	Previous    []*Token
	PreviousLen int
}

func NewCyclicAutomataStateBuilder(maxPrefixLength int) func() *CyclicAutomataState {
	return func() *CyclicAutomataState {
		return &CyclicAutomataState{
			Current:  make([]*Token, maxPrefixLength*2+1),
			Previous: make([]*Token, maxPrefixLength*2+1),
		}
	}
}

type CyclicParserStack struct {
	*ParserStack
	StatesLOS *ListOfStacks[CyclicAutomataState]
	State     CyclicAutomataState

	StatePool *Pool[CyclicAutomataState]
}

// NewCyclicParserStack creates a new CyclicParserStack initialized with one empty stack.
func NewCyclicParserStack(tokenStackPool *Pool[stack[*Token]], stateStackPool *Pool[stack[CyclicAutomataState]], statePool *Pool[CyclicAutomataState]) *CyclicParserStack {
	return &CyclicParserStack{
		ParserStack: NewParserStack(tokenStackPool),
		StatesLOS:   NewListOfStacks[CyclicAutomataState](stateStackPool),
		StatePool:   statePool,
	}
}

func (s *CyclicParserStack) Push(token *Token, state CyclicAutomataState) *Token {
	t := s.ParserStack.Push(token)

	st := s.StatePool.Get()
	copy(st.Current, state.Current)
	copy(st.Previous, state.Previous)
	st.CurrentLen = state.CurrentLen
	st.PreviousLen = state.PreviousLen

	s.StatesLOS.Push(*st)

	return t
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
	return s
}

func (s *CyclicParserStack) LastNonterminal() (*Token, error) {
	if s.State.CurrentLen >= 1 {
		return s.State.Current[0], nil
	}

	return nil, fmt.Errorf("no token stack current")
}

func (s *CyclicParserStack) CombineLOS(l *ListOfStacks[Token]) *ListOfStacks[Token] {
	list := NewListOfStacks[Token](l.pool)

	for i := 0; i < len(s.State.Previous) && s.State.Previous[i] != nil; i++ {
		s.State.Previous[i].Precedence = PrecEmpty
		list.Push(*s.State.Previous[i])
	}

	for i := 0; i < len(s.State.Current) && s.State.Current[i] != nil; i++ {
		s.State.Current[i].Precedence = PrecEmpty
		list.Push(*s.State.Current[i])
	}

	return list
}
