package utils

import "sync"

type Resetter interface {
	Reset()
}

type TypedPool[T Resetter] struct {
	pool *sync.Pool
}

func NewTypedPool[T Resetter](newFunc func() T) *TypedPool[T] {
	return &TypedPool[T]{
		pool: &sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

func (p *TypedPool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *TypedPool[T]) Put(x T) {
	p.pool.Put(x)
}
