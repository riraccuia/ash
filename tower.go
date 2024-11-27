package ash

import (
	"sync/atomic"
	"unsafe"
)

type Tower struct {
	top    *Level
	bottom *Level
}

func NewTower() *Tower {
	bottom := &Level{
		_next: nil,
		id:    0,
	}
	return &Tower{
		top:    bottom,
		bottom: bottom,
	}
}

func (l *Tower) SwapNext(level int, old, new unsafe.Pointer) (swapped bool) {
	if level > l.top.id {
		panic("unexpected level")
	}
	nlev := l.bottom
	for nlev != nil {
		if nlev.id == level {
			return atomic.CompareAndSwapPointer(&nlev._next, old, new)
		}
		nlev = nlev.Up()
	}
	return
}

func (l *Tower) UpdateNext(level int, new unsafe.Pointer) {
	if level > l.top.id {
		panic("unexpected level")
	}
	nlev := l.bottom
	for nlev != nil {
		if nlev.id == level {
			atomic.StorePointer(&nlev._next, new)
			return
		}
		nlev = nlev.Up()
	}
}

func (l *Tower) GetLevel(level int) *Level {
	lev := l.bottom
	for lev != nil {
		if lev.id == level {
			return lev
		}
		lev = lev.Up()
	}
	return nil
}

func (l *Tower) NextPtr(forLevel int) unsafe.Pointer {
	nlev := l.bottom
	for nlev != nil {
		if nlev.id == forLevel {
			return nlev.NextPtr()
		}
		nlev = nlev.Up()
	}
	return nil
}

func (l *Tower) AddPtr(level int, p unsafe.Pointer) {
	if l.bottom.id == level {
		l.bottom._next = p
		return
	}

	for level > l.top.id {
		id := l.top.id + 1
		newLev := &Level{
			_next: nil,
			_down: l.top,
			id:    id,
		}
		if id == level {
			newLev._next = p
		}
		l.top._up = newLev
		l.top = newLev
		return
	}

	nlev := l.bottom
	for nlev != nil {
		if nlev.id == level {
			nlev._next = p
		}
		nlev = nlev.Up()
	}
}
