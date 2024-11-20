package ash

import (
	"sync/atomic"
	"unsafe"
)

type tower struct {
	top    *level
	bottom *level
}

func NewTower() *tower {
	bottom := &level{
		_next: nil,
		id:    0,
	}
	return &tower{
		top:    bottom,
		bottom: bottom,
	}
}

type level struct {
	_next unsafe.Pointer
	_up   *level
	_down *level
	id    int
}

func (l *level) up() *level {
	return l._up
}

func (l *level) down() *level {
	return l._down
}

func (l *level) next() unsafe.Pointer {
	return atomic.LoadPointer(&l._next)
}

func (l *tower) getLevel(level int) *level {
	lev := l.bottom
	for lev != nil {
		if lev.id == level {
			return lev
		}
		lev = lev.up()
	}
	return nil
}

func (l *tower) next(fromLevel int) unsafe.Pointer {
	nlev := l.bottom
	for nlev != nil {
		if nlev.id == fromLevel {
			return nlev.next()
		}
		nlev = nlev.up()
	}
	return nil
}

func (l *tower) add(toLevel int, p unsafe.Pointer) {
	if l.bottom.id == toLevel {
		l.bottom._next = p
		return
	}

	if toLevel > l.top.id {
		newLev := &level{
			_next: p,
			_down: l.top,
			id:    toLevel,
		}
		l.top._up = newLev
		l.top = newLev
		return
	}

	/*var (
	plev *level
			nlev = l.bottom
		)

		for nlev != nil {
			if nlev.id == toLevel {
				nlev._next = p
				return
			}
			if toLevel > nlev.id {
				plev = nlev
				nlev = nlev.up()
				continue
			}
			newLev := &level{
				_next: p,
				_up:   nlev,
				_down: plev,
				id:    toLevel,
			}
			plev._up = newLev
			nlev._down = newLev
			return
		}*/
}

func (l *tower) swapNext(level int, old, new unsafe.Pointer) (swapped bool) {
	if level > l.top.id {
		panic("unexpected level")
	}
	nlev := l.bottom
	for nlev != nil {
		if nlev.id == level {
			return atomic.CompareAndSwapPointer(&nlev._next, old, new)
		}
		nlev = nlev.up()
	}
	return
}

func (l *tower) updateNext(level int, new unsafe.Pointer) {
	if level > l.top.id {
		panic("unexpected level")
	}
	nlev := l.bottom
	for nlev != nil {
		if nlev.id == level {
			atomic.StorePointer(&nlev._next, new)
			return
		}
		nlev = nlev.up()
	}
}
