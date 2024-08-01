package generator

import (
	"fmt"
	"github.com/giornetta/gopapageno"
	"math"
	"slices"
	"strings"
)

type precedenceMatrix [][]gopapageno.Precedence

type precedenceMap map[string]map[string]gopapageno.Precedence

func newPrecedenceMap(terminals []string) precedenceMap {
	m := make(precedenceMap)

	// Initialize an empty matrix.
	for _, term := range terminals {
		m[term] = make(map[string]gopapageno.Precedence)

		for _, term2 := range terminals {
			m[term][term2] = gopapageno.PrecEmpty
		}
	}

	return m
}

func (m precedenceMap) computeEqualsPrecedence(s gopapageno.ParsingStrategy, rules []rule, terminals []string, nonterminals *set[string]) error {
	for _, r := range rules {
		// Equals
		for _, term1 := range terminals {
			for _, term2 := range terminals {
				if m[term1][term2] == gopapageno.PrecEquals {
					continue
				}

				// Check if the two terminals are Eq in precedence.
				i1 := slices.Index(r.RHS, term1)
				if i1 == -1 {
					continue
				}

				i2 := slices.Index(r.RHS[i1+1:], term2) + i1 + 1

				// Check if the two terminals are present in the production, in the given order.
				if i2 == i1 || i1 >= i2 {
					continue
				}

				// Check if there is a token between them that isn't a nonterm
				if i1 == i2-1 && !nonterminals.Contains(r.RHS[i2-1]) {
					continue
				}

				if m[term1][term2] != gopapageno.PrecEmpty {
					return fmt.Errorf("precedence conflict on terminals %s and %s (%v, %v)", term1, term2, m[term1][term2], gopapageno.PrecEquals)
				}

				m[term1][term2] = gopapageno.PrecEquals
			}
		}
	}

	return nil
}

func (m precedenceMap) computeTakesPrecedence(s gopapageno.ParsingStrategy, rules []rule, terminals []string, nonterminals *set[string], rts map[string]*set[string]) error {
	for _, rule := range rules {
		for _, term1 := range terminals {
			for _, term2 := range terminals {
				if m[term1][term2] == gopapageno.PrecTakes || m[term1][term2] == gopapageno.PrecEquals {
					continue
				}

				success := false

				// Check if term2 is in the rhs
				indices := allIndices(rule.RHS, term2)
				// TODO: Check if this needs to iterate backwards, I suppose it's fine this way.
				for _, i2 := range indices {
					if success {
						break
					}

					// If term2 has no nonterminal before it
					if i2 == 0 || !nonterminals.Contains(rule.RHS[i2-1]) {
						continue
					}

					if rts[rule.RHS[i2-1]].Contains(term1) {
						if m[term1][term2] != gopapageno.PrecEmpty {
							return fmt.Errorf("precedence conflict on terminals %s and %s (%v, %v)", term1, term2, m[term1][term2], gopapageno.PrecTakes)
						}

						m[term1][term2] = gopapageno.PrecTakes
						success = true
					}
				}
			}
		}
	}

	return nil
}

func allIndices[T comparable](slice []T, el T) []int {
	indices := make([]int, 0)
	for i := range slice {
		if slice[i] == el {
			indices = append(indices, i)
		}
	}
	return indices
}

func (m precedenceMap) computeYieldsPrecedence(s gopapageno.ParsingStrategy, rules []rule, terminals []string, nonterminals *set[string], lts map[string]*set[string]) error {
	for _, rule := range rules {
		for _, term1 := range terminals {
			for _, term2 := range terminals {
				if m[term1][term2] == gopapageno.PrecYields || m[term1][term2] == gopapageno.PrecEquals {
					continue
				}

				success := false

				// Check if term1 is in the rhs
				indices := allIndices(rule.RHS, term1)
				for _, i1 := range indices {
					if success {
						break
					}

					//i1 := slices.Index(rule.RHS, term1)
					//if i1 == -1 {
					//	continue
					//}

					// If term2 has no nonterminal after it
					if i1 == len(rule.RHS)-1 || !nonterminals.Contains(rule.RHS[i1+1]) {
						continue
					}

					if lts[rule.RHS[i1+1]].Contains(term2) {
						if m[term1][term2] != gopapageno.PrecEmpty {
							return fmt.Errorf("precedence conflict on terminals %s and %s (%v, %v)", term1, term2, m[term1][term2], gopapageno.PrecYields)
						}
						m[term1][term2] = gopapageno.PrecYields
						success = true
					}
				}
			}
		}
	}

	return nil
}

func (m precedenceMap) buildMatrix(terminals []string) (precedenceMatrix, error) {
	for _, terminal := range terminals {
		if terminal != "_TERM" {
			m["_TERM"][terminal] = gopapageno.PrecYields
			m[terminal]["_TERM"] = gopapageno.PrecTakes
		}
	}
	m["_TERM"]["_TERM"] = gopapageno.PrecEquals

	precMatrix := make([][]gopapageno.Precedence, len(terminals))
	for i, t1 := range terminals {
		precMatrix[i] = make([]gopapageno.Precedence, len(terminals))

		for j, t2 := range terminals {
			precMatrix[i][j] = m[t1][t2]
		}
	}

	return precMatrix, nil
}

func (p *grammarDescription) newPrecedenceMatrix(opts *Options) (matrix precedenceMatrix, err error) {
	terminals := p.terminals.Slice()
	if err := moveToFront(terminals, "_TERM"); err != nil {
		return nil, fmt.Errorf("could not move _TERM to front: %w", err)
	}

	// TODO: Remove this when refactoring AOPP matrix creation.
	if opts.Strategy == gopapageno.AOPP {
		matrix, err = p.newAssociativePrecedenceMatrix()
	} else {
		lts, rts := p.getTerminalSets()

		m := newPrecedenceMap(terminals)

		if err := m.computeEqualsPrecedence(opts.Strategy, p.rules, terminals, p.nonterminals); err != nil {
			return nil, err
		}

		if err := m.computeTakesPrecedence(opts.Strategy, p.rules, terminals, p.nonterminals, rts); err != nil {
			return nil, err
		}

		if err := m.computeYieldsPrecedence(opts.Strategy, p.rules, terminals, p.nonterminals, lts); err != nil {
			return nil, err
		}

		matrix, err = m.buildMatrix(terminals)
	}

	if err != nil {
		return nil, err
	}

	var sb strings.Builder

	sb.WriteString("\t")
	for _, t := range terminals {
		sb.WriteString(fmt.Sprintf("\t%s", t))
	}
	sb.WriteString("\n")

	for i, t1 := range terminals {
		for j, _ := range terminals {
			if j == 0 {
				sb.WriteString(fmt.Sprintf("%s\t", t1))
			}

			sb.WriteString(fmt.Sprintf("%s\t", matrix[i][j]))
		}
		sb.WriteString("\n")
	}

	opts.Logger.Print(sb.String())

	return matrix, err
}

type conflict struct {
	rule rule
	i    int
	j    int
}

func (p *grammarDescription) newAssociativePrecedenceMatrix() (precedenceMatrix, error) {
	m := make(map[string]map[string]map[gopapageno.Precedence][]conflict)
	nonOP := make([]conflict, 0)

	// Initialize an empty matrix.
	for _, term := range p.terminals.Iter {
		m[term] = make(map[string]map[gopapageno.Precedence][]conflict)

		for _, term2 := range p.terminals.Iter {
			m[term][term2] = make(map[gopapageno.Precedence][]conflict)

			m[term][term2][gopapageno.PrecEquals] = make([]conflict, 0)
			m[term][term2][gopapageno.PrecYields] = make([]conflict, 0)
			m[term][term2][gopapageno.PrecTakes] = make([]conflict, 0)
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
				m[token1][token2][gopapageno.PrecEquals] = append(m[token1][token2][gopapageno.PrecEquals], conflict{
					rule: rule,
					i:    i,
					j:    i + 1,
				})
			} else if p.nonterminals.Contains(token1) && p.terminals.Contains(token2) {
				for _, token := range rts[token1].Iter {
					m[token][token2][gopapageno.PrecTakes] = append(m[token][token2][gopapageno.PrecTakes], conflict{
						rule: rule,
						i:    i,
						j:    i + 1,
					})
				}
			} else if p.terminals.Contains(token1) && p.nonterminals.Contains(token2) {
				for _, token := range lts[token2].Iter {
					m[token1][token][gopapageno.PrecYields] = append(m[token1][token][gopapageno.PrecYields], conflict{
						rule: rule,
						i:    i,
						j:    i + 1,
					})
				}
			} else {
				nonOP = append(nonOP, conflict{rule: rule, i: i, j: i + 1})
			}
		}

		//Check trigrams
		for i := 0; i < len(rhs)-2; i++ {
			token1 := rhs[i]
			token2 := rhs[i+1]
			token3 := rhs[i+2]

			if p.terminals.Contains(token1) && p.nonterminals.Contains(token2) && p.terminals.Contains(token3) {
				m[token1][token3][gopapageno.PrecEquals] = append(m[token1][token3][gopapageno.PrecEquals], conflict{
					rule: rule,
					i:    i,
					j:    i + 2,
				})
			}
		}
	}

	//set precedence for #
	for _, terminal := range p.terminals.Iter {
		if terminal != "_TERM" {
			m["_TERM"][terminal][gopapageno.PrecYields] = append(m["_TERM"][terminal][gopapageno.PrecYields], conflict{
				rule: p.rules[0],
				i:    0,
				j:    0,
			})

			m[terminal]["_TERM"][gopapageno.PrecTakes] = append(m[terminal]["_TERM"][gopapageno.PrecTakes], conflict{
				rule: p.rules[0],
				i:    0,
				j:    0,
			})
		}
	}
	m["_TERM"]["_TERM"][gopapageno.PrecEquals] = append(m["_TERM"]["_TERM"][gopapageno.PrecEquals], conflict{
		rule: p.rules[0],
		i:    0,
		j:    0,
	})

	if len(nonOP) > 0 {
		return nil, fmt.Errorf("rule %s violates OP form: no two nonterminals may be adjacent", nonOP[0].rule)
	}

	terminals := p.terminals.Slice()
	if err := moveToFront(terminals, "_TERM"); err != nil {
		return nil, fmt.Errorf("could not move _TERM to front: %w", err)
	}

	precMatrix := make([][]gopapageno.Precedence, len(terminals))

	for i, term := range terminals {
		precMatrix[i] = make([]gopapageno.Precedence, len(terminals))
		for j, term2 := range terminals {
			conflicts := make(map[gopapageno.Precedence][]conflict)

			var prec gopapageno.Precedence

			if len(m[term][term2][gopapageno.PrecEquals]) > 0 {
				conflicts[gopapageno.PrecEquals] = m[term][term2][gopapageno.PrecEquals]
				prec = gopapageno.PrecEquals
			}

			if len(m[term][term2][gopapageno.PrecTakes]) > 0 {
				conflicts[gopapageno.PrecTakes] = m[term][term2][gopapageno.PrecTakes]
				prec = gopapageno.PrecTakes
			}

			if len(m[term][term2][gopapageno.PrecYields]) > 0 {
				conflicts[gopapageno.PrecYields] = m[term][term2][gopapageno.PrecYields]
				prec = gopapageno.PrecYields
			}

			// Handle conflicts.
			// If `n : n T n` is present, it might be an associative conflict.
			if len(conflicts) > 1 {
				ok := false
				for p, cc := range conflicts {
					if p == gopapageno.PrecEquals {
						return nil, fmt.Errorf("strong precedence conflict")
					}

					// This is NOT enough, but we can leave it as is for testing purposes

					for _, c := range cc {
						if len(c.rule.RHS) == 3 && c.rule.RHS[0] == c.rule.LHS && c.rule.RHS[2] == c.rule.LHS && slices.Contains(terminals, c.rule.RHS[1]) {
							precMatrix[i][j] = gopapageno.PrecAssociative
							ok = true
							break
						}
					}

					if ok {
						break
					}
				}

				if !ok {
					return nil, fmt.Errorf("conflicts...")
				}
			}

			if len(conflicts) == 1 {
				precMatrix[i][j] = prec
			}

		}
	}

	return precMatrix, nil
}

// getTerminalSets returns two maps mapping nonterminal tokens to possible terminal productions.
func (p *grammarDescription) getTerminalSets() (lts map[string]*set[string], rts map[string]*set[string]) {
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
