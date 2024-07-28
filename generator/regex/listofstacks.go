package regex

import (
	"fmt"
	"math"
)

/*
LOS is a list of stacks containing symbols.
When the current stack is full a new one is automatically obtained and linked to it.
*/
type LOS struct {
	head *stack
	cur  *stack
	len  int
	pool *stackPool
}

/*
iterator allows to iterate over a LOS, either forward or backward.
*/
type iterator struct {
	los *LOS
	cur *stack
	pos int
}

/*
newLos creates a new LOS initialized with one empty stack.
*/
func newLos(pool *stackPool) LOS {
	firstStack := pool.GetSync()
	return LOS{firstStack, firstStack, 0, pool}
}

/*
Push pushes a symbol in the LOS.
It returns a pointer to the pushed symbol.
*/
func (l *LOS) Push(sym *symbol) *symbol {
	curStack := l.cur

	//If the current stack is full obtain a new stack and set it as the current one
	if curStack.Tos >= _STACK_SIZE {
		if curStack.Next != nil {
			curStack = curStack.Next
		} else {
			newStack := l.pool.GetSync()
			curStack.Next = newStack
			newStack.Prev = curStack
			curStack = newStack
		}
		l.cur = curStack
	}

	//Copy the symbol in the current position
	curStack.Data[curStack.Tos] = *sym

	//Save the pointer to the pushed symbol
	symPtr := &curStack.Data[curStack.Tos]

	//Increment the current position
	curStack.Tos++

	//Increment the total length of the list
	l.len++

	//Return the pointer to the pushed symbol
	return symPtr
}

/*
Pop pops a symbol from the stack and returns a pointer to it.
*/
func (l *LOS) Pop() *symbol {
	curStack := l.cur

	//Decrement the current position
	curStack.Tos--

	//While the current stack is empty set the previous stack as the current one.
	//If there are no more stacks return nil
	for curStack.Tos < 0 {
		curStack.Tos = 0

		if curStack.Prev == nil {
			return nil
		}

		curStack = curStack.Prev
		l.cur = curStack

		curStack.Tos--
	}

	//Decrement the total length of the list
	l.len--

	//Return the pointer to the symbol
	return &curStack.Data[curStack.Tos]
}

/*
Merge merges a LOS to another by linking their stacks.
*/
func (l *LOS) Merge(l2 LOS) {
	l.cur.Next = l2.head
	l2.head.Prev = l.cur
	l.cur = l2.cur
	l.len += l2.len
}

/*
Split splits a LOS into a number of lists equal to numSplits,
which are returned as a slice of LOS.
If there are not at least numSplits stacks in the LOS it panics.
The original LOS should not be used after this operation.
*/
func (l *LOS) Split(numSplits int) []LOS {
	if numSplits > l.NumStacks() {
		panic(fmt.Sprintln("Cannot apply", numSplits, "splits on a LOS containing only", l.NumStacks(), "stacks."))
	}

	listsOfStacks := make([]LOS, numSplits)
	curList := 0

	numStacks := l.NumStacks()
	deltaStacks := float64(numStacks) / float64(numSplits)
	totAssignedStacks := 0
	remainder := float64(0)

	curStack := l.head

	for totAssignedStacks < numStacks {
		remainder += deltaStacks
		stacksToAssign := int(math.Floor(remainder))

		curStack.Prev = nil
		listsOfStacks[curList] = LOS{curStack, curStack, curStack.Tos, l.pool}

		for i := 1; i < stacksToAssign; i++ {
			curStack = curStack.Next
			listsOfStacks[curList].cur = curStack
			listsOfStacks[curList].len += curStack.Tos
		}
		nextStack := curStack.Next
		curStack.Next = nil
		curStack = nextStack

		remainder -= float64(stacksToAssign)
		totAssignedStacks += stacksToAssign

		curList++
	}

	return listsOfStacks
}

/*
Length returns the number of symbols contained in the LOS
*/
func (l *LOS) Length() int {
	return l.len
}

/*
NumStacks returns the number of stacks contained in the LOS
*/
func (l *LOS) NumStacks() int {
	i := 0

	curStack := l.head

	for curStack != nil {
		i++
		curStack = curStack.Next
	}

	return i
}

/*
Println prints the content of the LOS.
*/
func (l *LOS) Println() {
	iterator := l.HeadIterator()

	sym := iterator.Next()
	for sym != nil {
		fmt.Printf("(%s, %s) -> ", tokenToString(sym.Token), precToString(sym.Precedence))
		sym = iterator.Next()
	}
	fmt.Println()
}

/*
HeadIterator returns an iterator initialized to point before the first element of the list.
*/
func (l *LOS) HeadIterator() iterator {
	return iterator{l, l.head, -1}
}

/*
TailIterator returns an iterator initialized to point after the last element of the list.
*/
func (l *LOS) TailIterator() iterator {
	return iterator{l, l.cur, l.cur.Tos}
}

/*
Prev moves the iterator one position backward and returns a pointer to the current symbol.
It returns nil if it points before the first element of the list.
*/
func (i *iterator) Prev() *symbol {
	curStack := i.cur

	i.pos--

	for i.pos < 0 {
		i.pos = -1
		if curStack.Prev == nil {
			return nil
		}
		curStack = curStack.Prev
		i.cur = curStack
		i.pos = curStack.Tos - 1
	}

	return &curStack.Data[i.pos]
}

/*
Cur returns a pointer to the current symbol.
It returns nil if it points before the first element or after the last element of the list.
*/
func (i *iterator) Cur() *symbol {
	curStack := i.cur

	if i.pos < 0 || i.pos >= curStack.Tos {
		return nil
	}

	return &curStack.Data[i.pos]
}

/*
Next moves the iterator one position forward and returns a pointer to the current symbol.
It returns nil if it points after the last element of the list.
*/
func (i *iterator) Next() *symbol {
	curStack := i.cur

	i.pos++

	for i.pos >= curStack.Tos {
		i.pos = curStack.Tos
		if curStack.Next == nil {
			return nil
		}
		curStack = curStack.Next
		i.cur = curStack
		i.pos = 0
	}

	return &curStack.Data[i.pos]
}
