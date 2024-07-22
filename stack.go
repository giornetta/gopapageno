package gopapageno

import (
	"math"
	"reflect"
)

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

func (s *stack[T]) Push(t T) {
	if s.Tos >= s.Size {
		panic("calculations were wrong.")
	}

	s.Data[s.Tos] = t

	s.Tos++
}

func (s *stack[T]) Replace(t T) {
	if s.Tos == 0 {
		panic("calculations were wrong.")
	}

	s.Data[s.Tos-1] = t
}

type ValueConstructor[T any] func() T

func newStack[T any]() *stack[T] {
	stackLen := stackSize[T]()

	return &stack[T]{
		Data: make([]T, stackLen),
		Size: stackLen,
	}
}

func (s *stack[T]) Slice(from int, length int) []T {
	return s.Data[from : from+length]
}

func stackSize[T any]() int {
	typeSize := reflect.TypeFor[T]().Size()
	return 1024 * 1024 / int(typeSize)
}

func stacksCount[T any](src []byte, concurrency int, avgTokenLen int) int {
	return int(math.Ceil(float64(len(src)) / float64(avgTokenLen) / float64(concurrency) / float64(stackSize[T]())))
}

func newStackBuilder[T any](c ValueConstructor[T]) func() *stack[T] {
	typeSize := reflect.TypeFor[T]().Size()
	stackLen := 1024 * 1024 / typeSize

	return func() *stack[T] {
		s := &stack[T]{
			Data: make([]T, stackLen),
			Size: int(stackLen),
		}

		if c != nil {
			for i, _ := range s.Data {
				s.Data[i] = c()
			}
		}

		return s
	}
}
