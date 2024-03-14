package automata

import "slices"

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
		NumStates: len(bytes) + 1,
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

func (ss NFAStateSet[T]) EpsilonClosure() NFAStateSet[T] {
	closureMap := make(map[*NFAState[T]]struct{})
	var e struct{}

	for _, s := range ss {
		closure := s.EpsilonClosure()

		for _, curClosureState := range closure {
			closureMap[curClosureState] = e
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

func stateSetMove(stateSet []*NFAState, char int) []*NFAState {
	states := make([]*NFAState, 0)

	for _, curState := range stateSet {
		reachedStates := curState.Transitions[char]

		for _, curReachedState := range reachedStates {
			if !slices.Contains(states, curReachedState) {
				states = append(states, curReachedState)
			}
		}
	}

	return states
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
func (nfa *NFA) Unite(nfa2 NFA) {
	newInitial := NFAState{}
	newFinal := NFAState{}

	newInitial.AddTransition(_EPSILON, nfa.Initial)
	newInitial.AddTransition(_EPSILON, nfa2.Initial)

	nfa.Final.AddTransition(_EPSILON, &newFinal)
	nfa2.Final.AddTransition(_EPSILON, &newFinal)

	nfa.Initial = &newInitial
	nfa.Final = &newFinal

	nfa.NumStates += nfa2.NumStates + 2
}

// Operator *
func (nfa *NFA) KleeneStar() {
	newInitial := NFAState{}
	newFinal := NFAState{}

	newInitial.AddTransition(_EPSILON, nfa.Initial)
	newInitial.AddTransition(_EPSILON, &newFinal)

	nfa.Final.AddTransition(_EPSILON, nfa.Initial)
	nfa.Final.AddTransition(_EPSILON, &newFinal)

	nfa.Initial = &newInitial
	nfa.Final = &newFinal

	nfa.NumStates += 2
}

// Operator +
func (nfa *NFA) KleenePlus() {
	newInitial := NFAState{}
	newFinal := NFAState{}

	newInitial.AddTransition(_EPSILON, nfa.Initial)

	nfa.Final.AddTransition(_EPSILON, nfa.Initial)
	nfa.Final.AddTransition(_EPSILON, &newFinal)

	nfa.Initial = &newInitial
	nfa.Final = &newFinal

	nfa.NumStates += 2
}

// Operator ?
func (nfa *NFA) ZeroOrOne() {
	newInitial := NFAState{}
	newFinal := NFAState{}

	newInitial.AddTransition(_EPSILON, nfa.Initial)
	newInitial.AddTransition(_EPSILON, &newFinal)

	nfa.Final.AddTransition(_EPSILON, &newFinal)

	nfa.Initial = &newInitial
	nfa.Final = &newFinal

	nfa.NumStates += 2
}

func (nfa *NFA) AddAssociatedRule(ruleNum int) {
	finalState := nfa.Final

	if finalState.AssociatedRules == nil {
		finalState.AssociatedRules = []int{ruleNum}
	} else {
		finalState.AssociatedRules = append(finalState.AssociatedRules, ruleNum)
	}
}

func (nfa *NFA) ToDfa() DFA {
	genStates := make([]nfaStateSetPtr, 0)

	curDfaStateNum := 0

	initialDfaState := DFAState{}
	initialDfaState.Num = curDfaStateNum

	dfa := DFA{&initialDfaState, make([]*DFAState, 0), 1}

	genStates = append(genStates, nfaStateSetPtr{nfa.Initial.EpsilonClosure(), &initialDfaState})

	search := func(gStates []nfaStateSetPtr, stateSet []*NFAState) *nfaStateSetPtr {
		for _, curGState := range gStates {
			if len(curGState.StateSet) != len(stateSet) {
				continue
			}
			equal := true
			for i, _ := range curGState.StateSet {
				if curGState.StateSet[i] != stateSet[i] {
					equal = false
					break
				}
			}
			if equal {
				return &curGState
			}
		}
		return nil
	}

	nextStateToCheckPos := 0

	for nextStateToCheckPos < len(genStates) {
		curStateSet := genStates[nextStateToCheckPos].StateSet
		curDfaState := genStates[nextStateToCheckPos].Ptr

		//For each character
		for i := 1; i < 256; i++ {
			charStateSet := stateSetMove(curStateSet, i)
			epsilonClosure := stateSetEpsilonClosure(charStateSet)

			if len(epsilonClosure) == 0 {
				continue
			}

			foundStateSetPtr := search(genStates, epsilonClosure)

			if foundStateSetPtr != nil {
				curDfaState.Transitions[i] = foundStateSetPtr.Ptr
			} else {
				curDfaStateNum++
				newDfaState := DFAState{}
				newDfaState.Num = curDfaStateNum
				newDfaState.AssociatedRules = make([]int, 0)
				for _, curNfaState := range epsilonClosure {
					newDfaState.AssociatedRules = append(newDfaState.AssociatedRules, curNfaState.AssociatedRules...)
				}
				curDfaState.Transitions[i] = &newDfaState
				newStateSetPtr := nfaStateSetPtr{epsilonClosure, &newDfaState}

				genStates = append(genStates, newStateSetPtr)
			}
		}
		nextStateToCheckPos++
	}

	dfa.NumStates = len(genStates)

	for _, genState := range genStates {
		if slices.Contains(genState.StateSet, nfa.Final) {
			finalState := genState.Ptr
			finalState.IsFinal = true
			dfa.Final = append(dfa.Final, finalState)
		}
	}

	return dfa
}

type nfaStateSetPtr struct {
	StateSet []*NFAState
	Ptr      *DFAState
}

func (stateSet1 *nfaStateSetPtr) Equals(stateSet2 *nfaStateSetPtr) bool {
	if len(stateSet1.StateSet) != len(stateSet2.StateSet) {
		return false
	}
	for i, _ := range stateSet1.StateSet {
		if stateSet1.StateSet[i] != stateSet2.StateSet[i] {
			return false
		}
	}
	return true
}
