package main

import "fmt"

type Item = interface{}

type Pool struct {
	buf   chan Item
	alloc func() Item
	clean func(Item) Item
}

func NewPool(size int, alloc func() Item, clean func(Item) Item) *Pool {
	return &Pool{
		buf:   make(chan Item, size),
		alloc: alloc,
		clean: clean,
	}
}

func (p *Pool) Get() Item {
	select {
	case i := <-p.buf:
		if p.clean != nil {
			return p.clean(i)
		}
		return i
	default:
		return p.alloc()
	}
}

func (p *Pool) Put(x Item) {
	select {
	case p.buf <- x:
	default:

	}
}

func test() {
	p := NewPool(1024,
		func() interface{} {
			return make([]byte, 0, 10)
		},
		func(i interface{}) interface{} {
			return i.([]byte)[0]
		})
	fmt.Println(p)
}
