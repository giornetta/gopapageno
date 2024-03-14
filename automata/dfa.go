package automata

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
