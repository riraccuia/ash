package ash

import (
	"math"
	"sync/atomic"
	"unsafe"
)

const (
	PValue = 0.5 // p = 1/2
)

var probabilities [64]uint32

func init() {
	probability := 1.0

	for level := 0; level < 64; level++ {
		probabilities[level] = uint32(probability * float64(math.MaxUint32))
		probability *= PValue
	}
}

func randomHeight(maxLevel int) int {
	seed := fastrand()

	height := 1
	for height < maxLevel && seed <= probabilities[height] {
		height++
	}

	return height
}

type Tower struct {
	top    *Level
	bottom *Level
}

type Level struct {
	_next unsafe.Pointer
	_up   *Level
	_down *Level
	id    int
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

func (l *Level) Up() *Level {
	return l._up
}

func (l *Level) Down() *Level {
	return l._down
}

func (l *Level) NextPtr() unsafe.Pointer {
	return atomic.LoadPointer(&l._next)
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
