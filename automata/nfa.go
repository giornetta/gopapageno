package automata

const _EPSILON byte = 0

type NFA[T comparable] struct {
	Initial   *NFAState[T]
	Final     *NFAState[T]
	NumStates int
}

type NFAState[T comparable] struct {
	Transitions     map[T][]*NFAState[T]
	AssociatedRules []int
}

type NFAStateSet[T comparable] []*NFAState[T]

func NewNFAState[T comparable]() *NFAState[T] {
	return &NFAState[T]{
		Transitions:     make(map[T][]*NFAState[T]),
		AssociatedRules: make([]int, 0),
	}
}

func NewEmptyNFA[T comparable]() *NFA[T] {
	s := NewNFAState[T]()

	return &NFA[T]{
		Initial:   s,
		Final:     s,
		NumStates: 1,
	}
}

func NFAFrom[T comparable](e T) *NFA[T] {
	nfa := &NFA[T]{
		Initial:   NewNFAState[T](),
		Final:     NewNFAState[T](),
		NumStates: 2,
	}

	nfa.Initial.AddTransition(e, nfa.Final)

	return nfa
}

func NFAFromClass[T comparable](elements map[T]bool) *NFA[T] {
	nfa := &NFA[T]{
		Initial:   NewNFAState[T](),
		Final:     NewNFAState[T](),
		NumStates: 2,
	}

	for e, transition := range elements {
		if transition {
			nfa.Initial.AddTransition(e, nfa.Final)
		}
	}

	return nfa
}

func NFAFromSlice[T comparable](elements []T) *NFA[T] {
	nfa := &NFA[T]{
		Initial:   NewNFAState[T](),
		NumStates: len(elements) + 1,
	}

	curState := nfa.Initial

	for _, b := range elements {
		newState := NewNFAState[T]()
		curState.AddTransition(b, newState)
		curState = newState
	}

	nfa.Final = curState

	return nfa
}

// EpsilonClosure returns the set of states which are reachable from [s] on epsilon-transitions.
func (s *NFAState[T]) EpsilonClosure() []*NFAState[T] {
	closure := make([]*NFAState[T], 0)

	closure = append(closure, s)

	nextStateToCheckPos := 0

	// We will assume epsilon to be the zero-value of T
	var epsilon T

	for nextStateToCheckPos < len(closure) {
		curState := closure[nextStateToCheckPos]
		closure = append(closure, curState.Transitions[epsilon]...)
		nextStateToCheckPos++
	}

	return closure
}

// EpsilonClosure returns the set of states which are reachable from the states in [ss] on epsilon-transitions.
func (ss NFAStateSet[T]) EpsilonClosure() NFAStateSet[T] {
	closureMap := make(map[*NFAState[T]]struct{})
	var e struct{}

	for _, s := range ss {
		closure := s.EpsilonClosure()

		for _, closureState := range closure {
			closureMap[closureState] = e
		}
	}

	keys := make([]*NFAState[T], len(closureMap), len(closureMap))
	i := 0

	for k := range closureMap {
		keys[i] = k
		i++
	}

	return keys
}

func (ss NFAStateSet[T]) Move(input T) NFAStateSet[T] {
	statesMap := make(map[*NFAState[T]]struct{})
	var e struct{}

	for _, s := range ss {
		reachedStates := s.Transitions[input]

		for _, rs := range reachedStates {
			statesMap[rs] = e
		}
	}

	keys := make([]*NFAState[T], len(statesMap), len(statesMap))
	i := 0

	for k := range statesMap {
		keys[i] = k
		i++
	}

	return keys
}

// AddTransition adds a transition from [s] to other based on [b].
func (s *NFAState[T]) AddTransition(b T, other *NFAState[T]) {
	if s.Transitions[b] == nil {
		s.Transitions[b] = []*NFAState[T]{other}
		return
	}

	s.Transitions[b] = append(s.Transitions[b], other)
}

func (nfa *NFA[T]) Concatenate(other *NFA[T]) {
	nfa.Final = other.Initial
	nfa.Final = other.Final

	nfa.NumStates = nfa.NumStates + other.NumStates - 1
}

// Operator |
func (nfa *NFA[T]) Unite(other *NFA[T]) {
	newInitial := NewNFAState[T]()
	newFinal := NewNFAState[T]()

	var epsilon T

	newInitial.AddTransition(epsilon, nfa.Initial)
	newInitial.AddTransition(epsilon, other.Initial)

	nfa.Final.AddTransition(epsilon, newFinal)
	other.Final.AddTransition(epsilon, newFinal)

	nfa.Initial = newInitial
	nfa.Final = newFinal

	nfa.NumStates += other.NumStates + 2
}

// Operator *
func (nfa *NFA[T]) KleeneStar() {
	newInitial := NewNFAState[T]()
	newFinal := NewNFAState[T]()

	var epsilon T

	newInitial.AddTransition(epsilon, nfa.Initial)
	newInitial.AddTransition(epsilon, newFinal)

	nfa.Final.AddTransition(epsilon, nfa.Initial)
	nfa.Final.AddTransition(epsilon, newFinal)

	nfa.Initial = newInitial
	nfa.Final = newFinal

	nfa.NumStates += 2
}

// Operator +
func (nfa *NFA[T]) KleenePlus() {
	newInitial := NewNFAState[T]()
	newFinal := NewNFAState[T]()

	var epsilon T

	newInitial.AddTransition(epsilon, nfa.Initial)

	nfa.Final.AddTransition(epsilon, nfa.Initial)
	nfa.Final.AddTransition(epsilon, newFinal)

	nfa.Initial = newInitial
	nfa.Final = newFinal

	nfa.NumStates += 2
}

// Operator ?
func (nfa *NFA[T]) ZeroOrOne() {
	newInitial := NewNFAState[T]()
	newFinal := NewNFAState[T]()

	var epsilon T

	newInitial.AddTransition(epsilon, nfa.Initial)
	newInitial.AddTransition(epsilon, newFinal)

	nfa.Final.AddTransition(epsilon, newFinal)

	nfa.Initial = newInitial
	nfa.Final = newFinal

	nfa.NumStates += 2
}

func (nfa *NFA[T]) AddAssociatedRule(ruleNum int) {
	finalState := nfa.Final

	if finalState.AssociatedRules == nil {
		finalState.AssociatedRules = []int{ruleNum}
	} else {
		finalState.AssociatedRules = append(finalState.AssociatedRules, ruleNum)
	}
}
