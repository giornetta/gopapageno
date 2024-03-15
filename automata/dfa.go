package automata

import "slices"

// DFA is a Deterministic Finite Automaton.
type DFA struct {
	Initial   *DFAState
	Final     []*DFAState
	NumStates int
}

// DFAState represents a state of a [DFA].
type DFAState struct {
	Num             int
	Transitions     [256]*DFAState
	IsFinal         bool
	AssociatedRules []int
}

func DFAFromNFA[T comparable](nfa *NFA[T]) *DFA {
	genStates := make([]nfaStateSetPtr[T], 0)

	curDfaStateNum := 0

	initialDfaState := DFAState{}
	initialDfaState.Num = curDfaStateNum

	dfa := DFA{&initialDfaState, make([]*DFAState, 0), 1}

	genStates = append(genStates, nfaStateSetPtr[T]{nfa.Initial.EpsilonClosure(), &initialDfaState})

	search := func(gStates []nfaStateSetPtr[T], stateSet []*NFAState[T]) *nfaStateSetPtr[T] {
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
			curStateSet.Move(i)
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

type nfaStateSetPtr[T comparable] struct {
	StateSet NFAStateSet[T]
	Ptr      *DFAState
}

func (stateSet1 *nfaStateSetPtr[T]) Equals(stateSet2 *nfaStateSetPtr[T]) bool {
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

func (dfaState *DFAState) getStatesR(addedStates *[]*DFAState) {
	//The state was already added, return
	if (*addedStates)[dfaState.Num] != nil {
		return
	}
	(*addedStates)[dfaState.Num] = dfaState

	for _, nextState := range dfaState.Transitions {
		if nextState != nil {
			nextState.getStatesR(addedStates)
		}
	}
}

/*
GetState returns a slice containing all the states of the dfa.
The states are sorted by their state number.
*/
func (dfa *DFA) GetStates() []*DFAState {
	states := make([]*DFAState, dfa.NumStates)

	dfa.Initial.getStatesR(&states)

	return states
}

/*func (dfa *DFA) Check(str []byte) (bool, bool, uint16) {
	curState := dfa.Initial

	//fmt.Println(curState)

	for _, curChar := range str {
		curState = curState.Transitions[curChar]

		//fmt.Println(curState)

		if curState == nil {
			return false, false, 0
		}
	}

	if len(curState.AssociatedTokens) == 0 {
		return curState.IsFinal, false, 0
	}

	index := 0
	minRule := curState.AssociatedTokens[0].RuleNum

	for i := 1; i < len(curState.AssociatedTokens); i++ {
		if curState.AssociatedTokens[i].RuleNum < minRule {
			minRule = curState.AssociatedTokens[i].RuleNum
			index = i
		}
	}

	return curState.IsFinal, true, curState.AssociatedTokens[index].Token
}*/
