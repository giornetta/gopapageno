package gopapageno

import (
	"context"
	"fmt"
)

// parseCyclic implements COPP.
func (w *parserWorker) parseCyclic(ctx context.Context, stack *CyclicParserStack, tokens *ListOfStacks[Token], nextToken *Token, finalPass bool, resultCh chan<- parseResult, errCh chan<- error) {
	tokensIt := tokens.HeadIterator()

	prefix := make([]TokenType, w.parser.MaxRHSLength)
	prefixTokens := make([]*Token, w.parser.MaxRHSLength)

	// If the thread is the first, push a # onto the stack
	// Otherwise, push the first inputToken onto the stack
	if !finalPass {
		if w.id == 0 {
			stack.Push(&Token{
				Type:       TokenTerm,
				Precedence: PrecEmpty,
			})
		} else {
			t := tokensIt.Next()
			t.Precedence = PrecEmpty
			stack.Push(t)
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

		//Find the first terminal on the stack and get the precedence between it and the current tokens inputToken
		firstTerminal := stack.FirstTerminal()

		if !inputToken.Type.IsTerminal() {
			prec = PrecYields
		} else {
			if firstTerminal == nil {
				prec = w.parser.precedence(TokenTerm, inputToken.Type)
			} else {
				prec = w.parser.precedence(firstTerminal.Type, inputToken.Type)

				if prec == PrecEquals && firstTerminal.Precedence == PrecTakes {
					prec = PrecYields
				}

				if prec == PrecEquals && firstTerminal.Precedence == PrecEmpty {
					prec = PrecTakes
				}
			}
		}

		// If it yields precedence, PUSH the inputToken onto the stack with its precedence relation.
		if prec == PrecYields {
			t := inputToken
			if inputToken.Type.IsTerminal() {
				inputToken.Precedence = prec
				inputToken = stack.Push(inputToken)
			}

			// If the current construction is a single nonterminal.
			if stack.IsCurrentSingleNonterminal() {
				// Append input character to the current construction.
				stack.AppendStateToken(t)
			} else {
				// Otherwise, swap.
				stack.SwapState()
				stack.AppendStateToken(t)
			}

			inputToken = tokensIt.Next()
		} else if prec == PrecEquals {
			inputToken.Precedence = prec
			// If it is equals, it is probably a shift transition?
			if inputToken.Type == TokenTerm {
				stack.Push(inputToken)
				break
			}

			oldIndex := stack.State.CurrentIndex
			// If the current construction is a single nonterminal.
			if stack.IsCurrentSingleNonterminal() {
				stack.State.CurrentIndex = stack.State.PreviousIndex
				stack.State.CurrentLen += stack.State.PreviousLen
			}

			rhsTokens = rhsTokens[:0]
			rhsTokens = append(rhsTokens, stack.Current()...)
			rhs = rhs[:0]
			for i := range stack.State.CurrentLen {
				rhs = append(rhs, rhsTokens[i].Type)
			}

			lhs, ruleNum := w.parser.findMatch(rhs)
			if lhs != TokenEmpty && w.parser.Rules[ruleNum].Type != RuleSimple {
				lhsToken, err := w.match(rhs, rhsTokens, true)
				if err != nil {
					errCh <- fmt.Errorf("worker %d could not match: %v", w.id, err)
					return
				}

				// Reset state
				stack.StateTokenStack.Tos = stack.State.CurrentIndex
				stack.State.CurrentLen = 0
				stack.AppendStateToken(lhsToken)

				_ = oldIndex
			}

			stack.AppendStateToken(inputToken)

			// Replace the topmost token on the stack, keeping its state unchanged.
			_, s := stack.Pop2()
			stack.PushWithState(inputToken, *s)

			inputToken = tokensIt.Next()
		} else if prec == PrecTakes {
			//If there are no tokens yielding precedence on the stack, push inputToken onto the stack.
			//Otherwise, perform a reduction. (Reduction == Pop/Shift move?)
			if stack.YieldingPrecedence() == 0 {
				inputToken.Precedence = prec
				stack.Push(inputToken)

				if inputToken.Type != TokenTerm {
					stack.SwapState()
				}

				inputToken = tokensIt.Next()
			} else {
				var i int
				// Prefix is made of a single nonterminal
				prefixTokens = prefixTokens[:0]
				prefix = prefix[:0]

				if stack.IsCurrentSingleNonterminal() {
					prefixTokens = append(prefixTokens, stack.Previous()...)
					for i = 0; i < stack.State.PreviousLen; i++ {
						prefix = append(prefix, prefixTokens[i].Type)
					}
				}

				prefixTokens = append(prefixTokens, stack.Current()...)
				for j := 0; j < stack.State.CurrentLen; j++ {
					prefix = append(prefix, prefixTokens[i].Type)

					i++
				}

				_, st := stack.Pop2()
				stack.UpdateFirstTerminal()

				// Prefix is made of a single nonterminal
				if st.CurrentLen == 1 && !stack.StateTokenStack.Data[st.CurrentIndex].IsTerminal() {
					stack.State.PreviousIndex = st.PreviousIndex
					stack.State.PreviousLen = st.PreviousLen
				} else {
					stack.State.PreviousIndex = st.CurrentIndex
					stack.State.PreviousLen = st.CurrentLen
				}

				rhsTokens = prefixTokens[:i]
				rhs = prefix[:i]

				lhsToken, err := w.match(rhs, rhsTokens, false)
				if err != nil {
					errCh <- fmt.Errorf("worker %d could not match: %v", w.id, err)
					return
				}

				// Reset state
				stack.StateTokenStack.Tos = stack.State.PreviousIndex + stack.State.PreviousLen + 1
				stack.State.CurrentIndex = stack.StateTokenStack.Tos - 1
				stack.State.CurrentLen = 1
				stack.StateTokenStack.Replace(lhsToken)
			}
		} else {
			//If there's no precedence relation, abort the parsing
			errCh <- fmt.Errorf("no precedence relation found")
			return
		}
	}

	resultCh <- parseResult{w.id, stack}
}

func (w *parserWorker) match(rhs []TokenType, rhsTokens []*Token, isPrefix bool) (*Token, error) {
	lhs, ruleNum := w.parser.findMatch(rhs)
	if lhs == TokenEmpty {
		return nil, fmt.Errorf("could not find match for rhs %v", rhs)
	}

	lhsToken := rhsTokens[0]

	ruleType := w.parser.Rules[ruleNum].Type
	if ruleType == RuleSimple || ruleType == RuleCyclic {
		lhsToken = w.ntPool.Get()
		lhsToken.Type = lhs
	}

	//Execute the semantic action
	w.parser.Func(ruleNum, lhsToken, rhsTokens, w.id)

	return lhsToken, nil
}

func (w *parserWorker) getNonterminal(rhsTokens []*Token) *Token {
	// Try to find the token associated to the leftmost token.
	lhsToken, ok := w.producedTokens[rhsTokens[0]]
	if !ok {
		lhsToken = w.ntPool.Get()
	}

	return lhsToken
}
