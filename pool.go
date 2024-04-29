package gopapageno

type Constructor[T any] func() *T

// A Pool can be used to preallocate a number of items of type T.
// It is not thread-safe.
type Pool[T any] struct {
	pool []T
	cur  int

	constructor Constructor[T]
}

type PoolOpt[T any] func(p *Pool[T])

func WithConstructor[T any](constructor Constructor[T]) PoolOpt[T] {
	return func(p *Pool[T]) {
		p.constructor = constructor

		if p.constructor != nil {
			for i := 0; i < len(p.pool); i++ {
				p.pool[i] = *p.constructor()
			}
		}
	}
}

// NewPool creates a new pool, allocating `length` elements.
func NewPool[T any](length int, opts ...PoolOpt[T]) *Pool[T] {
	p := &Pool[T]{
		pool:        make([]T, length),
		cur:         0,
		constructor: nil,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Get returns an item from the pool if available. Otherwise, it initializes a new one.
// It is not thread-safe.
func (p *Pool[T]) Get() *T {
	if p.cur >= len(p.pool) {
		if p.constructor == nil {
			return new(T)
		}

		return p.constructor()
	}

	addr := &p.pool[p.cur]
	p.cur++

	return addr
}

// Left returns the number of items remaining in the pool.
func (p *Pool[T]) Left() int {
	return len(p.pool) - p.cur
}
