package ash

import (
	"sync/atomic"
	"unsafe"
)

type layers [MaxLevel]unsafe.Pointer

func (lvl *layers) next(fromLevel int) unsafe.Pointer {
	return atomic.LoadPointer(&lvl[fromLevel])
}

func (lvl *layers) add(toLevel int, p unsafe.Pointer) {
	atomic.StorePointer(&lvl[toLevel], p)
}

func (lvl *layers) swapNext(level int, old, new unsafe.Pointer) (swapped bool) {
	return atomic.CompareAndSwapPointer(&lvl[level], old, new)
}

func (lvl *layers) updateNext(level int, new unsafe.Pointer) {
	atomic.StorePointer(&lvl[level], new)
}
