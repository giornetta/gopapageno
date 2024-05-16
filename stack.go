package gopapageno

import (
	"fmt"
	"math"
)

type ParserStack struct {
	*ListOfStacks[*Token]

	firstTerminal *Token
	yieldsPrec    int
}

// NewParserStack creates a new ParserStack initialized with one empty stack.
func NewParserStack(pool *Pool[stack[*Token]]) *ParserStack {
	return &ParserStack{
		ListOfStacks: NewListOfStacks[*Token](pool),
	}
}

// FirstTerminal returns a pointer to the first terminal on the stack.
func (s *ParserStack) FirstTerminal() *Token {
	return s.firstTerminal
}

// UpdateFirstTerminal should be used after a reduction in order to update the first terminal counter.
// In fact, in order to save some time, only the Push operation automatically updates the first terminal pointer,
// while the Pop operation does not.
func (s *ParserStack) UpdateFirstTerminal() {
	s.firstTerminal = s.findFirstTerminal()
}

// findFirstTerminals computes the first terminal on the stacks.
// This function is for internal usage only.
func (s *ParserStack) findFirstTerminal() *Token {
	curStack := s.cur

	pos := curStack.Tos - 1

	for pos < 0 {
		pos = -1
		if curStack.Prev == nil {
			return nil
		}
		curStack = curStack.Prev
		pos = curStack.Tos - 1
	}

	for !curStack.Data[pos].Type.IsTerminal() {
		pos--
		for pos < 0 {
			pos = -1
			if curStack.Prev == nil {
				return nil
			}
			curStack = curStack.Prev
			pos = curStack.Tos - 1
		}
	}

	return curStack.Data[pos]
}

// Push pushes a token pointer in the ParserStack.
// It returns the pointer itself.
func (s *ParserStack) Push(token *Token) *Token {
	// If the current stack is full, we must obtain a new one and set it as the current one.
	if s.cur.Tos >= s.cur.Size {
		if s.cur.Next != nil {
			s.cur = s.cur.Next
		} else {
			stack := s.pool.Get()

			s.cur.Next = stack
			stack.Prev = s.cur

			s.cur = stack
		}
	}

	s.cur.Data[s.cur.Tos] = token

	//If the token is a terminal update the firstTerminal pointer
	if token.Type.IsTerminal() {
		s.firstTerminal = token
	}

	// If the token is yielding precedence, increase the counter
	if token.Precedence == PrecYields || token.Precedence == PrecAssociative {
		s.yieldsPrec++
	}

	s.cur.Tos++
	s.len++

	return token
}

// Pop removes the topmost element from the ParserStack and returns it.
func (s *ParserStack) Pop() *Token {
	s.cur.Tos--

	if s.cur.Tos < 0 {
		s.cur.Tos = 0

		if s.cur.Prev == nil {
			return nil
		}

		s.cur = s.cur.Prev
		s.cur.Tos--
	}

	t := s.cur.Data[s.cur.Tos]
	if t.Precedence == PrecYields || t.Precedence == PrecAssociative {
		s.yieldsPrec--
	}

	s.len--

	return t
}

func (s *ParserStack) YieldingPrecedence() int {
	return s.yieldsPrec
}

// Merge links the stacks of the current and of another ParserStack.
func (s *ParserStack) Merge(other *ParserStack) {
	s.ListOfStacks.Merge(other.ListOfStacks)

	s.firstTerminal = other.firstTerminal
}

// Split splits a listOfStacks into a slice of listOfStacks of length n.
// The original listOfStacks should not be used after this operation.
func (s *ParserStack) Split(n int) ([]*ParserStack, error) {
	numStacks := s.NumStacks()

	if n > numStacks {
		return nil, fmt.Errorf("not enough stacks in ParserStack")
	}

	lists := make([]*ParserStack, n)
	curList := 0

	deltaStacks := float64(numStacks) / float64(n)
	assignedStacks := 0
	remainder := float64(0)

	curStack := s.head

	for assignedStacks < numStacks {
		remainder += deltaStacks

		stacksToAssign := int(math.Floor(remainder + 0.5))

		curStack.Prev = nil
		lists[curList] = &ParserStack{
			ListOfStacks: &ListOfStacks[*Token]{
				head: curStack,
				cur:  curStack,
				len:  curStack.Tos,
				pool: s.pool,
			},
		}

		for i := 1; i < stacksToAssign; i++ {
			curStack = curStack.Next
			lists[curList].cur = curStack
			lists[curList].len += curStack.Tos
		}

		next := curStack.Next
		curStack.Next = nil

		curStack = next

		lists[curList].firstTerminal = lists[curList].findFirstTerminal()

		remainder -= float64(stacksToAssign)
		assignedStacks += stacksToAssign

		curList++
	}

	return lists, nil
}

func (s *ParserStack) Combine(o Stacker) Stacker {
	var topLeft Token

	// TODO: This could be moved in Push/Pop to allow constant time access.
	it := s.HeadIterator()
	for t := it.Next(); t != nil && t.Precedence != PrecYields; t = it.Next() {
		topLeft = *t
	}

	list := NewParserStack(s.pool)

	topLeft.Precedence = PrecEmpty
	list.Push(&topLeft)

	for t := it.Cur(); t != nil && t.Precedence != PrecTakes; t = it.Next() {
		list.Push(t)
	}

	list.UpdateFirstTerminal()

	return list
}

func (s *ParserStack) CombineNoAlloc() {
	var topLeft *Token

	var topLeftStack *stack[*Token]
	var topLeftPos int

	// TODO: This could be moved in Push/Pop to allow constant time access.
	it := s.HeadIterator()
	removedTokens := 0
	for t := it.Next(); t != nil && t.Precedence != PrecYields; t = it.Next() {
		topLeft = t
		topLeftStack = it.cur
		topLeftPos = it.pos

		removedTokens++
	}

	topLeft.Precedence = PrecEmpty

	s.cur = topLeftStack
	s.len -= removedTokens
	s.cur.Tos = topLeftPos
}

func (s *ParserStack) CombineLOS(l *ListOfStacks[Token]) *ListOfStacks[Token] {
	var tok Token

	list := NewListOfStacks[Token](l.pool)

	it := s.HeadIterator()

	// Ignore first element
	it.Next()

	for t := it.Next(); t != nil && t.Precedence != PrecYields; t = it.Next() {
		tok = *t
		tok.Precedence = PrecEmpty
		list.Push(tok)
	}

	return list
}

func (s *ParserStack) LastNonterminal() (*Token, error) {
	for token := s.Pop(); token != nil; token = s.Pop() {
		if !token.Type.IsTerminal() {
			return token, nil
		}
	}

	return nil, fmt.Errorf("no nonterminal found")
}

// ParserStackIterator allows to iterate over a listOfTokenPointerStacks, either forward or backward.
type ParserStackIterator struct {
	los *ParserStack

	cur *stack[*Token]
	pos int
}

// HeadIterator returns an iterator initialized to point before the first element of the list.
func (s *ParserStack) HeadIterator() *ParserStackIterator {
	return &ParserStackIterator{s, s.head, -1}
}

// TailIterator returns an iterator initialized to point after the last element of the list.
func (s *ParserStack) TailIterator() *ParserStackIterator {
	return &ParserStackIterator{s, s.cur, s.cur.Tos}
}

// Prev moves the iterator one position backward and returns a pointer to the new current element.
// It returns nil if it points before the first element of the list.
func (i *ParserStackIterator) Prev() *Token {
	curStack := i.cur

	i.pos--

	if i.pos >= 0 {
		return curStack.Data[i.pos]
	}

	i.pos = -1
	if curStack.Prev == nil {
		return nil
	}
	curStack = curStack.Prev
	i.cur = curStack
	i.pos = curStack.Tos - 1

	return curStack.Data[i.pos]
}

// Cur returns a pointer to the current element.
// It returns nil if it points before the first element or after the last element of the list.
func (i *ParserStackIterator) Cur() *Token {
	curStack := i.cur

	if i.pos >= 0 && i.pos < curStack.Tos {
		return curStack.Data[i.pos]
	}

	return nil
}

// Next moves the iterator one position forward and returns a pointer to the new current element.
// It returns nil if it points after the last element of the list.
func (i *ParserStackIterator) Next() *Token {
	curStack := i.cur

	i.pos++

	if i.pos < curStack.Tos {
		return curStack.Data[i.pos]
	}

	i.pos = curStack.Tos
	if curStack.Next == nil {
		return nil
	}
	curStack = curStack.Next
	i.cur = curStack
	i.pos = 0

	return curStack.Data[i.pos]
}
