package main

import "sync"

type resetter interface {
	Reset()
}

type Pool[T resetter] struct {
	pool sync.Pool
}

func NewPool[T resetter]() *Pool[T] {
	return &Pool[T]{}
}

func (p *Pool[T]) Get() T {
	v, _ := p.pool.Get().(T)
	return v
}

func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
