package ash

import (
	"sync/atomic"
	"unsafe"
)

type Tower [CapLevel]unsafe.Pointer

func (lvl *Tower) NextPtr(fromLevel int) unsafe.Pointer {
	return atomic.LoadPointer(&lvl[fromLevel])
}

func (lvl *Tower) AddPtrUnsafe(toLevel int, p unsafe.Pointer) {
	lvl[toLevel] = p
}

func (lvl *Tower) AddPtr(toLevel int, p unsafe.Pointer) {
	atomic.StorePointer(&lvl[toLevel], p)
}

func (lvl *Tower) SwapNext(level int, old, new unsafe.Pointer) (swapped bool) {
	return atomic.CompareAndSwapPointer(&lvl[level], old, new)
}

func (lvl *Tower) UpdateNext(level int, new unsafe.Pointer) {
	atomic.StorePointer(&lvl[level], new)
}
