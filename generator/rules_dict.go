package generator

// rulesDictionary is a data structure used to store unique RHS -> LHS mappings.
// It is useful to remove repeated RHS productions.
type rulesDictionary struct {
	KeysRHS [][]string

	ValuesLHS  []*set[string]
	SemActions []*string
}

func newRulesDictionary(capacity int) *rulesDictionary {
	return &rulesDictionary{
		KeysRHS:    make([][]string, 0, capacity),
		ValuesLHS:  make([]*set[string], 0, capacity),
		SemActions: make([]*string, 0, capacity)}
}

func (d *rulesDictionary) Add(r *rule) {
	found := false

	for i, keyRhs := range d.KeysRHS {
		if rhsEquals(keyRhs, r.RHS) {
			d.ValuesLHS[i].Add(r.LHS)
			found = true

			break
		}
	}

	if !found {
		d.KeysRHS = append(d.KeysRHS, r.RHS)

		d.ValuesLHS = append(d.ValuesLHS, newSet[string]())
		d.ValuesLHS[len(d.ValuesLHS)-1].Add(r.LHS)

		d.SemActions = append(d.SemActions, &r.Action)
	}
}

func (d *rulesDictionary) Remove(rhs []string) {
	for i, curKeyRHS := range d.KeysRHS {
		if rhsEquals(curKeyRHS, rhs) {
			d.KeysRHS = append(d.KeysRHS[:i], d.KeysRHS[i+1:]...)
			d.ValuesLHS = append(d.ValuesLHS[:i], d.ValuesLHS[i+1:]...)
			d.SemActions = append(d.SemActions[:i], d.SemActions[i+1:]...)
		}
	}
}

func (d *rulesDictionary) Len() int {
	return len(d.KeysRHS)
}

func (d *rulesDictionary) Copy() *rulesDictionary {
	newDict := newRulesDictionary(d.Len())

	for i, _ := range d.KeysRHS {
		newDict.KeysRHS = append(newDict.KeysRHS, make([]string, len(d.KeysRHS[i])))
		copy(newDict.KeysRHS[i], d.KeysRHS[i])

		newDict.ValuesLHS = append(newDict.ValuesLHS, d.ValuesLHS[i].Copy())

		newDict.SemActions = append(newDict.SemActions, d.SemActions[i])
	}

	return newDict
}

func (d *rulesDictionary) LHSSets() []*set[string] {
	valueLHSSets := make([]*set[string], 0)

	for _, curLHSSet := range d.ValuesLHS {
		alreadyContained := false
		for _, curSet := range valueLHSSets {
			if curSet.Equals(curLHSSet) {
				alreadyContained = true
				break
			}
		}
		if !alreadyContained {
			valueLHSSets = append(valueLHSSets, curLHSSet)
		}
	}

	return valueLHSSets
}
