package gopapageno

import (
	"context"
	"fmt"
	"slices"
)

// parseCyclic implements COPP.
func (w *parserWorker) parseCyclic(ctx context.Context, stack *CyclicParserStack, tokens *ListOfStacks[Token], nextToken *Token, finalPass bool, resultCh chan<- parseResult, errCh chan<- error) {
	tokensIt := tokens.HeadIterator()

	curRhsPrefixLen := 0
	prevRhsPrefixLen := 0

	state := CyclicAutomataState{
		Current:  make([]*Token, w.parser.MaxPrefixLength*2+1),
		Previous: make([]*Token, w.parser.MaxPrefixLength*2+1),
	}

	prefixTokens := make([]TokenType, w.parser.MaxPrefixLength*2+1)
	prefix := make([]*Token, w.parser.MaxPrefixLength*2+1)

	// If the thread is the first, push a # onto the stack
	// Otherwise, push the first inputToken onto the stack
	if !finalPass {
		if w.id == 0 {
			stack.Push(&Token{
				Type:       TokenTerm,
				Value:      nil,
				Precedence: PrecEmpty,
				Next:       nil,
				Child:      nil,
			}, state)
		} else {
			t := tokensIt.Next()
			t.Precedence = PrecEmpty

			// TODO: State should probably be changed once I implement parallel parsing.
			stack.Push(t, state)
		}

		// If the thread is the last, push a # onto the tokens m
		// Otherwise, push onto the tokens m the first inputToken of the next tokens m
		if w.id == w.parser.concurrency-1 {
			tokens.Push(Token{
				Type:       TokenTerm,
				Value:      nil,
				Precedence: PrecEmpty,
				Next:       nil,
				Child:      nil,
			})
		} else if nextToken != nil {
			tokens.Push(*nextToken)
		}
	}

	var lhsToken *Token

	var rhs []TokenType
	var rhsTokens []*Token

	newNonTerm := Token{
		Type:       TokenEmpty,
		Value:      nil,
		Precedence: PrecEmpty,
		Next:       nil,
		Child:      nil,
	}

	// Iterate over the tokens
	// If this is the first worker, start reading from the input stack, otherwise begin with the last
	// token of the previous stack.
	for inputToken := tokensIt.Next(); inputToken != nil; {
		//If the current inputToken is a non-terminal, push it onto the stack with no precedence relation
		if !inputToken.Type.IsTerminal() {
			inputToken.Precedence = PrecEmpty
			stack.Push(inputToken, state)

			inputToken = tokensIt.Next()
			continue
		}

		//Find the first terminal on the stack and get the precedence between it and the current tokens inputToken
		firstTerminal := stack.FirstTerminal()

		var prec Precedence
		if firstTerminal == nil {
			prec = w.parser.precedence(TokenTerm, inputToken.Type)
		} else {
			prec = w.parser.precedence(firstTerminal.Type, inputToken.Type)
		}

		// If it yields precedence, PUSH the inputToken onto the stack with its precedence relation.
		if prec == PrecYields {
			inputToken.Precedence = prec
			t := stack.Push(inputToken, state)

			// If the current construction is a single nonterminal.
			if curRhsPrefixLen == 1 && !state.Current[0].Type.IsTerminal() {
				// Append input character to the current construction.
				state.Current[curRhsPrefixLen] = t
				curRhsPrefixLen++
			} else {
				// Otherwise, swap.
				copy(state.Previous, state.Current)
				prevRhsPrefixLen = curRhsPrefixLen

				for i := 1; i < curRhsPrefixLen; i++ {
					state.Current[i] = nil
				}
				state.Current[0] = t
				curRhsPrefixLen = 1
			}

			inputToken = tokensIt.Next()
		} else if prec == PrecEquals {
			inputToken.Precedence = prec
			// If it is equals, it is probably a shift transition?

			// If the current construction is a single nonterminal.
			if curRhsPrefixLen == 1 && !state.Current[0].Type.IsTerminal() {
				// Prepend previous construction to current one; leaving the previous one untouched.
				copy(state.Current[prevRhsPrefixLen:], state.Current)
				copy(state.Current[:prevRhsPrefixLen], state.Previous)

				curRhsPrefixLen += prevRhsPrefixLen
			}

			// Append input character to the current construction.
			state.Current[curRhsPrefixLen] = inputToken
			curRhsPrefixLen++

			for i := range curRhsPrefixLen {
				prefixTokens[i] = state.Current[i].Type
			}

			// If the construction has a suffix which is a double occurrence of a string produced by a Kleene-+.
			for _, prefix := range w.parser.Prefixes {
				if slices.Equal(prefix, prefixTokens[:curRhsPrefixLen]) {
					// Try this out: parse as rhs all tokens of the prefix except last one, and substitute the resulting lhs to them.
					rhsTokens = state.Current[:curRhsPrefixLen-1]
					rhs = prefix[:curRhsPrefixLen-1]

					lhs, ruleNum := w.parser.findMatch(rhs)
					if lhs == TokenEmpty {
						errCh <- fmt.Errorf("could not find match for rhs %v", rhs)
						return
					}

					lhs = rhs[0]

					newNonTerm.Type = lhs
					lhsToken = w.ntPool.Get()
					*lhsToken = newNonTerm

					//Execute the semantic action
					w.parser.Func(ruleNum, lhsToken, rhsTokens, w.id)

					// Reset state
					for i := 1; i < len(state.Current); i++ {
						state.Current[i] = nil
					}
					state.Current[0] = lhsToken
					state.Current[1] = inputToken
					curRhsPrefixLen = 2

					break
				}
			}

			// Replace the topmost token on the stack, setting its precedence to Yield.
			_, s := stack.Pop2()
			inputToken.Precedence = PrecYields
			stack.Push(inputToken, *s)

			inputToken = tokensIt.Next()
		} else if prec == PrecTakes {
			//If there are no tokens yielding precedence on the stack, push inputToken onto the stack.
			//Otherwise, perform a reduction. (Reduction == Pop/Shift move?)
			if stack.YieldingPrecedence() == 0 {
				tok := state.Current[0]

				inputToken.Precedence = prec
				state.Current[0] = inputToken

				stack.Push(tok, state)

				inputToken = tokensIt.Next()
			} else {
				_, st := stack.Pop2()

				var i int
				// Prefix is made of a single nonterminal
				if curRhsPrefixLen == 1 && !state.Current[0].Type.IsTerminal() {
					for i = 0; i < prevRhsPrefixLen && state.Previous[i] != nil; i++ {
						prefixTokens[i] = state.Previous[i].Type
						prefix[i] = state.Previous[i]
					}
				}

				for j := 0; j < curRhsPrefixLen && state.Current[j] != nil; j++ {
					prefixTokens[i] = state.Current[j].Type
					prefix[i] = state.Current[j]

					i++
				}

				stack.UpdateFirstTerminal()

				// Prefix is made of a single nonterminal
				if w.parser.MaxPrefixLength > 0 && st.Current[0] != nil && !st.Current[0].Type.IsTerminal() && st.Current[1] == nil {
					state.Previous = st.Previous
				} else {
					state.Previous = st.Current
				}

				rhsTokens = prefix[:i]
				rhs = prefixTokens[:i]

				lhs, ruleNum := w.parser.findMatch(rhs)
				if lhs == TokenEmpty {
					errCh <- fmt.Errorf("could not find match for rhs %v", rhs)
					return
				}

				newNonTerm.Type = lhs
				lhsToken = w.ntPool.Get()
				*lhsToken = newNonTerm

				//Execute the semantic action
				w.parser.Func(ruleNum, lhsToken, rhsTokens, w.id)

				// Reset state
				for i := 1; i < len(state.Current); i++ {
					state.Current[i] = nil
				}
				state.Current[0] = lhsToken
				curRhsPrefixLen = 1

				i = 0
				for _, t := range state.Previous {
					if t != nil {
						i++
					}
				}
				prevRhsPrefixLen = i
			}
		} else {
			//If there's no precedence relation, abort the parsing
			errCh <- fmt.Errorf("no precedence relation found")
			return
		}
	}

	resultCh <- parseResult{w.id, stack}
}
