package gopapageno

type ParserStackBase[T Tokener] struct {
	*LOPS[T]

	firstTerminal *T
}

// NewParserStack creates a new ParserStackBase initialized with one empty stack.
func NewParserStack[T Tokener](pool *Pool[stack[*T]]) *ParserStackBase[T] {
	return &ParserStackBase[T]{
		LOPS: NewLOPS[T](pool),
	}
}

// FirstTerminal returns a pointer to the first terminal on the stack.
func (s *ParserStackBase[T]) FirstTerminal() *T {
	return s.firstTerminal
}

// UpdateFirstTerminal should be used after a reduction in order to update the first terminal counter.
// In fact, in order to save some time, only the Push operation automatically updates the first terminal pointer,
// while the Pop operation does not.
func (s *ParserStackBase[T]) UpdateFirstTerminal() {
	s.firstTerminal = s.findFirstTerminal()
}

// findFirstTerminals computes the first terminal on the stacks.
// This function is for internal usage only.
func (s *ParserStackBase[T]) findFirstTerminal() *T {
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

	for !(*curStack.Data[pos]).IsTerminal() {
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

// Push pushes a token pointer in the ParserStackBase.
// It returns the pointer itself.
func (s *ParserStackBase[T]) Push(token *T) *T {
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
	if (*token).IsTerminal() {
		s.firstTerminal = token
	}

	s.cur.Tos++
	s.len++

	return token
}

// Pop removes the topmost element from the ParserStackBase and returns it.
func (s *ParserStackBase[T]) Pop() *T {
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
	s.len--

	return t
}

// ParserStackIterator allows to iterate over a listOfTokenPointerStacks, either forward or backward.
type ParserStackIterator[T Tokener] struct {
	stack *ParserStackBase[T]

	cur *stack[*T]
	pos int
}

// HeadIterator returns an iterator initialized to point before the first element of the list.
func (s *ParserStackBase[T]) HeadIterator() *ParserStackIterator[T] {
	return &ParserStackIterator[T]{s, s.head, s.headFirst - 1}
}

// TailIterator returns an iterator initialized to point after the last element of the list.
func (s *ParserStackBase[T]) TailIterator() *ParserStackIterator[T] {
	return &ParserStackIterator[T]{s, s.cur, s.cur.Tos}
}

// Prev moves the iterator one position backward and returns a pointer to the new current element.
// It returns nil if it points before the first element of the list.
func (i *ParserStackIterator[T]) Prev() *T {
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
func (i *ParserStackIterator[T]) Cur() *T {
	curStack := i.cur

	if i.pos >= 0 && i.pos < curStack.Tos {
		return curStack.Data[i.pos]
	}

	return nil
}

// Next moves the iterator one position forward and returns a pointer to the new current element.
// It returns nil if it points after the last element of the list.
func (i *ParserStackIterator[T]) Next() *T {
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

func (i *ParserStackIterator[T]) IsLast() bool {
	if i.pos+1 < i.cur.Tos {
		return false
	}

	if i.cur.Next == nil {
		return true
	}

	return false
}
