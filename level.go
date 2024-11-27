package ash

import (
	"sync/atomic"
	"unsafe"
)

type Level struct {
	_next unsafe.Pointer
	_up   *Level
	_down *Level
	id    int
}

func (l *Level) Up() *Level {
	return l._up
}

func (l *Level) Down() *Level {
	return l._down
}

func (l *Level) NextPtr() unsafe.Pointer {
	return atomic.LoadPointer(&l._next)
}
