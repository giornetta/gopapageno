package gopapageno

import (
	"context"
	"fmt"
	"slices"
)

// parseCyclic implements COPP.
func (w *parserWorker) parseCyclic(ctx context.Context, stack *CyclicParserStack, tokens *ListOfStacks[Token], nextToken *Token, finalPass bool, resultCh chan<- parseResult, errCh chan<- error) {
	tokensIt := tokens.HeadIterator()

	var state CyclicAutomataState
	if stack.State.Current == nil {
		state = CyclicAutomataState{
			Current:  make([]*Token, 0, w.parser.MaxRHSLength),
			Previous: make([]*Token, 0, w.parser.MaxRHSLength),
		}
	} else {
		state = *stack.State
	}

	prefixTokens := make([]TokenType, w.parser.MaxRHSLength)
	prefix := make([]*Token, w.parser.MaxRHSLength)

	// If the thread is the first, push a # onto the stack
	// Otherwise, push the first inputToken onto the stack
	if !finalPass {
		if w.id == 0 {
			stack.Push(&Token{
				Type:       TokenTerm,
				Precedence: PrecEmpty,
			}, state)
		} else {
			t := tokensIt.Next()
			t.Precedence = PrecEmpty
			stack.Push(t, state)
		}

		// If the thread is the last, push a # onto the tokens m
		// Otherwise, push onto the tokens m the first inputToken of the next tokens m
		if w.id == w.parser.concurrency-1 {
			tokens.Push(Token{
				Type:       TokenTerm,
				Precedence: PrecEmpty,
			})
		} else if nextToken != nil {
			tokens.Push(*nextToken)
		}
	}

	var rhs []TokenType
	var rhsTokens []*Token

	// Iterate over the tokens
	// If this is the first worker, start reading from the input stack, otherwise begin with the last
	// token of the previous stack.
	for inputToken := tokensIt.Next(); inputToken != nil; {
		//If the current inputToken is a non-terminal, push it onto the stack with no precedence relation
		var prec Precedence

		pushTakes := false
		if !inputToken.Type.IsTerminal() {
			prec = PrecYields
		} else {
			//Find the first terminal on the stack and get the precedence between it and the current tokens inputToken
			firstTerminal := stack.FirstTerminal()

			if firstTerminal == nil {
				prec = w.parser.precedence(TokenTerm, inputToken.Type)
			} else {
				prec = w.parser.precedence(firstTerminal.Type, inputToken.Type)

				if firstTerminal.Precedence == PrecEmpty && inputToken.Type != TokenTerm && prec == PrecTakes {
					prec = PrecYields
					pushTakes = true
				}

				if prec == PrecEquals && (firstTerminal.Precedence == PrecTakes || firstTerminal.Precedence == PrecEmpty) {
					prec = PrecYields
				}
			}
		}

		// If it yields precedence, PUSH the inputToken onto the stack with its precedence relation.
		if prec == PrecYields {
			t := inputToken
			if inputToken.Type.IsTerminal() {
				if pushTakes {
					inputToken.Precedence = PrecTakes
				} else {
					inputToken.Precedence = prec
				}

				inputToken = stack.Push(inputToken, state)
			}

			// If the current construction is a single nonterminal.
			if len(state.Current) == 1 && !state.Current[0].Type.IsTerminal() {
				// Append input character to the current construction.
				state.Current = append(state.Current, t)
			} else {
				// Otherwise, swap.
				state.Previous = state.Previous[:0]
				state.Previous = append(state.Previous, state.Current...)

				state.Current = state.Current[:0]
				state.Current = append(state.Current, t)
			}

			inputToken = tokensIt.Next()
		} else if prec == PrecEquals {
			inputToken.Precedence = prec
			// If it is equals, it is probably a shift transition?
			if inputToken.Type == TokenTerm {
				stack.Push(inputToken, state)
				break
			}

			// If the current construction is a single nonterminal.
			if len(state.Current) == 1 && !state.Current[0].Type.IsTerminal() {
				// Prepend previous construction to current one; leaving the previous one untouched.
				state.Current = state.Current[:len(state.Current)+len(state.Previous)]
				if len(state.Current)+len(state.Previous) > cap(state.Current) {
					newCurrent := make([]*Token, 0, cap(state.Current)*2)
					newCurrent = append(newCurrent, state.Current...)
					state.Current = newCurrent
				}

				copy(state.Current[len(state.Previous):], state.Current[:len(state.Current)-len(state.Previous)])
				copy(state.Current, state.Previous)
			}

			// Append input character to the current construction.
			state.Current = append(state.Current, inputToken)

			prefixTokens = prefixTokens[:0]
			for i := range len(state.Current) {
				prefixTokens = append(prefixTokens, state.Current[i].Type)
			}

			// If the construction has a suffix which is a double occurrence of a string produced by a Kleene-+.
			for _, prefix := range w.parser.Prefixes {
				if slices.Equal(prefix, prefixTokens[:len(state.Current)]) {
					// Try this out: parse as rhs all tokens of the prefix except last one, and substitute the resulting lhs to them.
					rhsTokens = state.Current[:len(state.Current)-1]
					rhs = prefix[:len(state.Current)-1]

					lhsToken, err := w.match(rhs, rhsTokens, true)
					if err != nil {
						errCh <- fmt.Errorf("worker %d could not match prefix: %v", w.id, err)
						return
					}

					// Reset state
					state.Current = state.Current[:2]
					state.Current[0] = lhsToken
					state.Current[1] = inputToken

					break
				}
			}

			// Replace the topmost token on the stack, setting its precedence to Yield.
			t, s := stack.Pop2()
			inputToken.Precedence = t.Precedence
			stack.Push(inputToken, *s)

			inputToken = tokensIt.Next()
		} else if prec == PrecTakes {
			//If there are no tokens yielding precedence on the stack, push inputToken onto the stack.
			//Otherwise, perform a reduction. (Reduction == Pop/Shift move?)
			if stack.YieldingPrecedence() == 0 {
				inputToken.Precedence = prec
				stack.Push(inputToken, state)

				state.Current = append(state.Current, inputToken)

				inputToken = tokensIt.Next()
			} else {
				var i int
				// Prefix is made of a single nonterminal
				prefix = prefix[:0]
				prefixTokens = prefixTokens[:0]

				if len(state.Current) == 1 && !state.Current[0].Type.IsTerminal() {
					for i = 0; i < len(state.Previous); i++ {
						prefixTokens = append(prefixTokens, state.Previous[i].Type)
						prefix = append(prefix, state.Previous[i])
					}
				}

				for j := 0; j < len(state.Current); j++ {
					prefixTokens = append(prefixTokens, state.Current[j].Type)
					prefix = append(prefix, state.Current[j])

					i++
				}

				_, st := stack.Pop2()
				stack.UpdateFirstTerminal()

				// Prefix is made of a single nonterminal
				state.Previous = state.Previous[:0]
				if len(st.Current) == 1 && !st.Current[0].Type.IsTerminal() {
					state.Previous = append(state.Previous, st.Previous...)
				} else {
					state.Previous = append(state.Previous, st.Current...)
				}

				rhsTokens = prefix[:i]
				rhs = prefixTokens[:i]

				lhsToken, err := w.match(rhs, rhsTokens, false)
				if err != nil {
					errCh <- fmt.Errorf("worker %d could not match: %v", w.id, err)
					return
				}

				// Reset state
				state.Current = state.Current[:0]
				state.Current = append(state.Current, lhsToken)
			}
		} else {
			//If there's no precedence relation, abort the parsing
			errCh <- fmt.Errorf("no precedence relation found")
			return
		}
	}

	stack.State = &state

	resultCh <- parseResult{w.id, stack}
}

func (w *parserWorker) match(rhs []TokenType, rhsTokens []*Token, isPrefix bool) (*Token, error) {
	lhs, ruleNum := w.parser.findMatch(rhs)
	if lhs == TokenEmpty {
		return nil, fmt.Errorf("could not find match for rhs %v", rhs)
	}

	if isPrefix {
		lhs = rhs[0]
	}

	lhsToken := w.ntPool.Get()
	lhsToken.Type = lhs

	//Execute the semantic action
	w.parser.Func(ruleNum, lhsToken, rhsTokens, w.id)

	return lhsToken, nil
}
