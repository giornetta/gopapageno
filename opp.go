package gopapageno

import (
	"context"
	"fmt"
)

type parserWorker struct {
	parser *Parser

	id int

	ntPool *Pool[Token]
}

func (w *parserWorker) parse(ctx context.Context, stack Stacker, tokens *ListOfStacks[Token], nextToken *Token, finalPass bool, resultCh chan<- parseResult, errCh chan<- error) {
	switch s := stack.(type) {
	case *ParserStack:
		w.parseAcyclic(ctx, s, tokens, nextToken, finalPass, resultCh, errCh)
	case *CyclicParserStack:
		w.parseCyclic(ctx, s, tokens, nextToken, finalPass, resultCh, errCh)
	default:
		panic("unreachable")
	}
}

// parseAcyclic implements both OPP and AOPP strategies.
func (w *parserWorker) parseAcyclic(ctx context.Context, stack *ParserStack, tokens *ListOfStacks[Token], nextToken *Token, finalPass bool, resultCh chan<- parseResult, errCh chan<- error) {
	tokensIt := tokens.HeadIterator()

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
				Value:      nil,
				Precedence: PrecEmpty,
				Next:       nil,
				Child:      nil,
			})
		} else if nextToken != nil {
			tokens.Push(*nextToken)
		}
	}

	var pos int
	var lhsToken *Token

	var rhs []TokenType
	var rhsTokens []*Token

	rhsBuf := make([]TokenType, w.parser.MaxRHSLength)
	rhsTokensBuf := make([]*Token, w.parser.MaxRHSLength)

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
			stack.Push(inputToken)

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

		// If it's equal in precedence or yields, push the inputToken onto the stack with its precedence relation.
		if prec == PrecEquals || prec == PrecYields {
			inputToken.Precedence = prec
			stack.Push(inputToken)

			inputToken = tokensIt.Next()
		} else if prec == PrecTakes || prec == PrecAssociative {
			//If there are no tokens yielding precedence on the stack, push inputToken onto the stack.
			//Otherwise, perform a reduction
			if stack.YieldingPrecedence() == 0 {
				inputToken.Precedence = prec
				stack.Push(inputToken)

				inputToken = tokensIt.Next()
			} else {
				pos = w.parser.MaxRHSLength - 1

				var token *Token
				// Pop tokens from the stack until one that yields precedence is reached, saving them in rhsBuf
				for token = stack.Pop(); token.Precedence != PrecYields && token.Precedence != PrecAssociative; token = stack.Pop() {
					rhsTokensBuf[pos] = token
					rhsBuf[pos] = token.Type
					pos--
				}

				rhsTokensBuf[pos] = token
				rhsBuf[pos] = token.Type

				//Pop one last token, if it's a non-terminal add it to rhsBuf, otherwise ignore it (push it again onto the stack)
				token = stack.Pop()
				if token.Type.IsTerminal() {
					stack.Push(token)
				} else {
					pos--
					rhsTokensBuf[pos] = token
					rhsBuf[pos] = token.Type

					stack.UpdateFirstTerminal()
				}

				//Obtain the actual rhs from the buffers
				rhsTokens = rhsTokensBuf[pos:]
				rhs = rhsBuf[pos:]

				//Find corresponding lhs and ruleNum
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

				//Push the new nonterminal onto the stack
				stack.Push(lhsToken)
			}
		} else {
			//If there's no precedence relation, abort the parsing
			errCh <- fmt.Errorf("no precedence relation found")
			return
		}
	}

	resultCh <- parseResult{w.id, stack}
}
