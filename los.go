package gopapageno

import (
	"fmt"
	"math"
	"reflect"
)

// This is 1MB per stack (on 64 bit architecture)
const stackSize int = 26214

// This is 1MB per stack (on 64 bit architecture)
const pointerStackSize int = 131072

// stack contains a fixed size array of items,
// the current position in the stack and pointers to the previous and next stacks.
type stack[T any] struct {
	// Data [stackSize]T
	Data []T
	Tos  int
	Size int

	Prev *stack[T]
	Next *stack[T]
}

func newStack[T any]() *stack[T] {
	typeSize := reflect.TypeFor[T]().Size()
	stackLen := 1024 * 1024 / typeSize
	return &stack[T]{
		Data: make([]T, stackLen),
		Size: int(stackLen),
	}
}

// ListOfStacks is a list of stacks.
type ListOfStacks[T any] struct {
	head *stack[T]
	cur  *stack[T]

	len  int
	pool *Pool[stack[T]]
}

// LosIterator allows to iterate over a ListOfStacks, either forward or backwards.
type LosIterator[T any] struct {
	los *ListOfStacks[T]

	cur *stack[T]
	pos int
}

// NewListOfStacks creates a new ListOfStacks initialized with an empty stack.
func NewListOfStacks[T any](pool *Pool[stack[T]]) *ListOfStacks[T] {
	s := pool.Get()

	return &ListOfStacks[T]{
		head: s,
		cur:  s,
		len:  0,
		pool: pool,
	}
}

// Push adds an element to the listOfStacks.
// By default, the element is added to the current stack;
// if that is full, a new one is obtained from the pool.
func (l *ListOfStacks[T]) Push(t T) *T {
	// If the current stack is full, we must obtain a new one and set it as the current one.
	if l.cur.Tos >= l.cur.Size {
		if l.cur.Next != nil {
			l.cur = l.cur.Next
		} else {
			s := l.pool.Get()

			l.cur.Next = s
			s.Prev = l.cur

			l.cur = s
		}
	}

	l.cur.Data[l.cur.Tos] = t
	ptr := &l.cur.Data[l.cur.Tos]

	l.cur.Tos++
	l.len++

	return ptr
}

// Pop removes the topmost element from the listOfStacks and returns it.
func (l *ListOfStacks[T]) Pop() *T {
	l.cur.Tos--

	if l.cur.Tos >= 0 {
		l.len--
		return &l.cur.Data[l.cur.Tos]
	}

	l.cur.Tos = 0

	if l.cur.Prev == nil {
		return nil
	}

	l.cur = l.cur.Prev
	l.cur.Tos--
	l.len--

	return &l.cur.Data[l.cur.Tos]
}

func CombineLOS(l *ListOfStacks[Token], stacks *ParserStack) *ListOfStacks[Token] {
	var tok Token

	list := NewListOfStacks[Token](l.pool)

	it := stacks.HeadIterator()

	// Ignore first element
	it.Next()

	for t := it.Next(); t != nil && t.Precedence != PrecYields; t = it.Next() {
		tok = *t
		tok.Precedence = PrecEmpty
		list.Push(tok)
	}

	return list
}

// Get returns the topmost element from the ListOfStacks.
func (l *ListOfStacks[T]) Get() *T {
	if l.cur.Tos > 0 {
		return &l.cur.Data[l.cur.Tos-1]
	}

	if l.cur.Prev == nil {
		return nil
	}

	return &l.cur.Prev.Data[l.cur.Prev.Tos-1]
}

// Clear empties the ListOfStacks.
func (l *ListOfStacks[T]) Clear() {
	// Reset length
	l.len = 0

	// Reset Top of Stack for every stack
	for s := l.head; s != nil; s = s.Next {
		s.Tos = 0
	}

	// Reset current stack
	l.cur = l.head
}

// Merge links the stacks of the current and of another listOfStacks.
func (l *ListOfStacks[T]) Merge(other *ListOfStacks[T]) {
	l.cur.Next = other.head
	other.head.Prev = l.cur

	l.cur = other.cur
	l.len += other.len
}

// Split splits a listOfStacks into a slice of listOfStacks of length n.
// The original listOfStacks should not be used after this operation.
func (l *ListOfStacks[T]) Split(n int) ([]*ListOfStacks[T], error) {
	numStacks := l.NumStacks()

	if n > numStacks {
		return nil, fmt.Errorf("not enough stacks in listOfStacks")
	}

	lists := make([]*ListOfStacks[T], n)
	curList := 0

	deltaStacks := float64(numStacks) / float64(n)
	assignedStacks := 0
	remainder := float64(0)

	curStack := l.head

	for assignedStacks < numStacks {
		remainder += deltaStacks

		stacksToAssign := int(math.Floor(remainder + 0.5))

		curStack.Prev = nil
		lists[curList] = &ListOfStacks[T]{
			head: curStack,
			cur:  curStack,
			len:  curStack.Tos,
			pool: l.pool,
		}

		for i := 1; i < stacksToAssign; i++ {
			curStack = curStack.Next
			lists[curList].cur = curStack
			lists[curList].len += curStack.Tos
		}

		next := curStack.Next
		curStack.Next = nil

		curStack = next

		remainder -= float64(stacksToAssign)
		assignedStacks += stacksToAssign

		curList++
	}

	return lists, nil
}

// NumStacks returns the number of stacks contained in the listOfStacks.
// It takes linear time (in the number of stacks) to execute.
func (l *ListOfStacks[T]) NumStacks() int {
	n := 0

	for cur := l.head; cur != nil; cur = cur.Next {
		n++
	}
	return n
}

// Length returns the number of items contained in the listOfStacks.
// It takes constant time to execute.
func (l *ListOfStacks[T]) Length() int {
	return l.len
}

// HeadIterator returns an iterator initialized to point before the first element of the list.
func (l *ListOfStacks[T]) HeadIterator() *LosIterator[T] {
	return &LosIterator[T]{l, l.head, -1}
}

// TailIterator returns an iterator initialized to point after the last element of the list.
func (l *ListOfStacks[T]) TailIterator() *LosIterator[T] {
	return &LosIterator[T]{l, l.cur, l.cur.Tos}
}

// Prev moves the iterator one position backward and returns a pointer to the new current element.
// It returns nil if it points before the first element of the list.
func (i *LosIterator[T]) Prev() *T {
	curStack := i.cur

	i.pos--

	if i.pos >= 0 {
		return &curStack.Data[i.pos]
	}

	i.pos = -1
	if curStack.Prev == nil {
		return nil
	}
	curStack = curStack.Prev
	i.cur = curStack
	i.pos = curStack.Tos - 1

	return &curStack.Data[i.pos]
}

// Cur returns a pointer to the current element.
// It returns nil if it points before the first element or after the last element of the list.
func (i *LosIterator[T]) Cur() *T {
	curStack := i.cur

	if i.pos >= 0 && i.pos < curStack.Tos {
		return &curStack.Data[i.pos]
	}

	return nil
}

// Next moves the iterator one position forward and returns a pointer to the new current element.
// It returns nil if it points after the last element of the list.
func (i *LosIterator[T]) Next() *T {
	curStack := i.cur

	i.pos++

	if i.pos < curStack.Tos {
		return &curStack.Data[i.pos]
	}

	i.pos = curStack.Tos
	if curStack.Next == nil {
		return nil
	}
	curStack = curStack.Next
	i.cur = curStack
	i.pos = 0

	return &curStack.Data[i.pos]
}
