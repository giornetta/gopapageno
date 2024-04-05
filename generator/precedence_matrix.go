package generator

import (
	"fmt"
	"github.com/giornetta/gopapageno"
	"math"
)

type precedenceMatrix [][]gopapageno.Precedence

func (p *parserDescriptor) newPrecedenceMatrix() (precedenceMatrix, error) {
	m := make(map[string]map[string]gopapageno.Precedence)

	// Initialize an empty matrix.
	for _, term := range p.terminals.Iter {
		m[term] = make(map[string]gopapageno.Precedence)

		for _, term2 := range p.terminals.Iter {
			m[term][term2] = gopapageno.PrecEmpty
		}
	}

	lts, rts := p.getTerminalSets()

	for _, rule := range p.rules {
		rhs := rule.RHS

		//Check digrams
		for i := 0; i < len(rhs)-1; i++ {
			token1 := rhs[i]
			token2 := rhs[i+1]

			if p.terminals.Contains(token1) && p.terminals.Contains(token2) {
				//Check if the matrix already contains an entry for this couple
				if m[token1][token2] != gopapageno.PrecEmpty && m[token1][token2] != gopapageno.PrecEquals {
					return nil, fmt.Errorf("the precedence relation Equals is not unique between %s and %s", token1, token2)
				}

				m[token1][token2] = gopapageno.PrecEquals
			} else if p.nonterminals.Contains(token1) && p.terminals.Contains(token2) {
				for _, token := range rts[token1].Iter {
					//Check if the matrix already contains an entry for this couple
					if m[token][token2] != gopapageno.PrecEmpty && m[token][token2] != gopapageno.PrecTakes {
						return nil, fmt.Errorf("the precedence relation Takes is not unique between %s and %s", token, token2)
					}
					m[token][token2] = gopapageno.PrecTakes
				}
			} else if p.terminals.Contains(token1) && p.nonterminals.Contains(token2) {
				for _, token := range lts[token2].Iter {
					//Check if the matrix already contains an entry for this couple
					if m[token1][token] != gopapageno.PrecEmpty && m[token1][token] != gopapageno.PrecYields {
						return nil, fmt.Errorf("the precedence relation Yields is not unique between %s and %s", token1, token)
					}
					m[token1][token] = gopapageno.PrecYields
				}
			} else {
				return nil, fmt.Errorf("the rule %s is not in operator precedence form", rule.String())
			}
		}

		//Check trigrams
		for i := 0; i < len(rhs)-2; i++ {
			token1 := rhs[i]
			token2 := rhs[i+1]
			token3 := rhs[i+2]

			if p.terminals.Contains(token1) && p.nonterminals.Contains(token2) && p.terminals.Contains(token3) {
				//Check if the matrix already contains an entry for this couple
				if m[token1][token3] != gopapageno.PrecEmpty && m[token1][token3] != gopapageno.PrecEquals {
					return nil, fmt.Errorf("the precedence relation is not unique between %s and %s", token1, token3)
				}

				m[token1][token3] = gopapageno.PrecEquals
			}
		}
	}

	//set precedence for #
	for _, terminal := range p.terminals.Iter {
		if terminal != "_TERM" {
			m["_TERM"][terminal] = gopapageno.PrecYields
			m[terminal]["_TERM"] = gopapageno.PrecTakes
		}
	}
	m["_TERM"]["_TERM"] = gopapageno.PrecEquals

	terminals := p.terminals.Slice()
	if err := moveToFront(terminals, "_TERM"); err != nil {
		return nil, fmt.Errorf("could not move _TERM to front: %w", err)
	}

	precMatrix := make([][]gopapageno.Precedence, len(terminals))
	for i, t1 := range terminals {
		precMatrix[i] = make([]gopapageno.Precedence, len(terminals))

		for j, t2 := range terminals {
			precMatrix[i][j] = m[t1][t2]
		}
	}

	return precMatrix, nil
}

// getTerminalSets returns two maps mapping nonterminal tokens to possible terminal productions.
func (p *parserDescriptor) getTerminalSets() (lts map[string]*set[string], rts map[string]*set[string]) {
	lts = make(map[string]*set[string], p.nonterminals.Len())
	rts = make(map[string]*set[string], p.nonterminals.Len())

	// Initialize empty sets for every nonterminal token.
	for _, nonterminal := range p.nonterminals.Iter {
		lts[nonterminal] = newSet[string]()
		rts[nonterminal] = newSet[string]()
	}

	// Direct terminals.
	// If a terminal token is found in any RHS, add it to the corresponding LHS' sets.
	for _, rule := range p.rules {
		for i := 0; i < len(rule.RHS); i++ {
			token := rule.RHS[i]
			if p.terminals.Contains(token) {
				lts[rule.LHS].Add(token)
				break
			}
		}

		for i := len(rule.RHS) - 1; i >= 0; i-- {
			token := rule.RHS[i]
			if p.terminals.Contains(token) {
				rts[rule.LHS].Add(token)
				break
			}
		}
	}

	// Indirect terminals.
	// Loop until a fixed point is found.
	modified := true
	for modified {
		modified = false

		for _, rule := range p.rules {
			lhs := rule.LHS
			rhs := rule.RHS

			// If the first token on the RHS is a nonterminal,
			// Try adding every terminal token produced by firstToken directly
			// to the tokens produced by the considered lhs nonterminal.
			firstToken := rhs[0]
			if p.nonterminals.Contains(firstToken) {
				for _, token := range lts[firstToken].Iter {
					if !lts[lhs].Contains(token) {
						lts[lhs].Add(token)
						modified = true
					}
				}
			}

			// Do the same for the last token of the RHS.
			lastToken := rhs[len(rhs)-1]
			if p.nonterminals.Contains(lastToken) {
				for _, token := range rts[lastToken].Iter {
					if !rts[lhs].Contains(token) {
						rts[lhs].Add(token)
						modified = true
					}
				}
			}
		}
	}

	return lts, rts
}

// bitPack packs the matrix into a slice of uint64 where a precedence value is represented by just 2 bits.
func bitPack(matrix [][]gopapageno.Precedence) []uint64 {
	newSize := int(math.Ceil(float64(len(matrix)*len(matrix)) / float64(32)))

	newMatrix := make([]uint64, newSize)

	setPrec := func(elem *uint64, pos uint, prec gopapageno.Precedence) {
		bitMask := uint64(0x3 << pos)
		*elem = (*elem & ^bitMask) | (uint64(prec) << pos)
	}

	for i, _ := range matrix {
		for j, prec := range matrix[i] {
			flatElemPos := i*len(matrix) + j
			newElemPtr := &newMatrix[flatElemPos/32]
			newElemPos := uint((flatElemPos % 32) * 2)
			setPrec(newElemPtr, newElemPos, prec)
		}
	}

	return newMatrix
}

func moveToFront[T comparable](slice []T, e T) error {
	index := -1

	for i, v := range slice {
		if v == e {
			index = i
		}
	}

	if index == -1 {
		return fmt.Errorf("could not find element %v in given slice", e)
	}

	newSlice := append(slice[:index], slice[index+1:]...)
	newSlice = append([]T{e}, newSlice...)

	copy(slice, newSlice)

	return nil
}

func (p *parserDescriptor) newAssociativePrecedenceMatrix() (precedenceMatrix, error) {
	m := make(map[string]map[string]gopapageno.Precedence)

	// Initialize an empty matrix.
	for _, term := range p.terminals.Iter {
		m[term] = make(map[string]gopapageno.Precedence)

		for _, term2 := range p.terminals.Iter {
			m[term][term2] = gopapageno.PrecEmpty
		}
	}

	lts, rts := p.getTerminalSets()

	for _, rule := range p.rules {
		rhs := rule.RHS

		//Check digrams
		for i := 0; i < len(rhs)-1; i++ {
			token1 := rhs[i]
			token2 := rhs[i+1]

			if p.terminals.Contains(token1) && p.terminals.Contains(token2) {
				// Check if the matrix already contains an entry for this couple
				if m[token1][token2] != gopapageno.PrecEmpty && m[token1][token2] != gopapageno.PrecEquals {
					return nil, fmt.Errorf("the precedence relation Equals is not unique between %s and %s", token1, token2)
				}

				m[token1][token2] = gopapageno.PrecEquals
			} else if p.nonterminals.Contains(token1) && p.terminals.Contains(token2) {
				for _, token := range rts[token1].Iter {
					// Check if the matrix already contains an entry for this couple
					if m[token][token2] != gopapageno.PrecEmpty && m[token][token2] != gopapageno.PrecTakes {
						// TODO: Check for Weak or Associativity Conflict
						return nil, fmt.Errorf("the precedence relation Takes is not unique between %s and %s", token, token2)
					}
					m[token][token2] = gopapageno.PrecTakes
				}
			} else if p.terminals.Contains(token1) && p.nonterminals.Contains(token2) {
				for _, token := range lts[token2].Iter {
					// Check if the matrix already contains an entry for this couple
					if m[token1][token] != gopapageno.PrecEmpty && m[token1][token] != gopapageno.PrecYields {
						// TODO: Check for Weak or Associativity Conflict
						return nil, fmt.Errorf("the precedence relation Yields is not unique between %s and %s", token1, token)
					}
					m[token1][token] = gopapageno.PrecYields
				}
			} else {
				return nil, fmt.Errorf("the rule %s is not in operator precedence form", rule.String())
			}
		}

		//Check trigrams
		for i := 0; i < len(rhs)-2; i++ {
			token1 := rhs[i]
			token2 := rhs[i+1]
			token3 := rhs[i+2]

			if p.terminals.Contains(token1) && p.nonterminals.Contains(token2) && p.terminals.Contains(token3) {
				//Check if the matrix already contains an entry for this couple
				if m[token1][token3] != gopapageno.PrecEmpty && m[token1][token3] != gopapageno.PrecEquals {
					return nil, fmt.Errorf("the precedence relation is not unique between %s and %s", token1, token3)
				}

				m[token1][token3] = gopapageno.PrecEquals
			}
		}
	}

	//set precedence for #
	for _, terminal := range p.terminals.Iter {
		if terminal != "_TERM" {
			m["_TERM"][terminal] = gopapageno.PrecYields
			m[terminal]["_TERM"] = gopapageno.PrecTakes
		}
	}
	m["_TERM"]["_TERM"] = gopapageno.PrecEquals

	terminals := p.terminals.Slice()
	if err := moveToFront(terminals, "_TERM"); err != nil {
		return nil, fmt.Errorf("could not move _TERM to front: %w", err)
	}

	precMatrix := make([][]gopapageno.Precedence, len(terminals))
	for i, t1 := range terminals {
		precMatrix[i] = make([]gopapageno.Precedence, len(terminals))

		for j, t2 := range terminals {
			precMatrix[i][j] = m[t1][t2]
		}
	}

	return precMatrix, nil
}
