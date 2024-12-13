package ash

import (
	"sync/atomic"
	"unsafe"
)

// Tower embeds the skip list's levels, all of its methods
// are safe for concurrent use unless otherwise specified.
type Tower [CapLevel]unsafe.Pointer

// NextPtr returns the next pointer for the given level.
func (lvl *Tower) NextPtr(forLevel int) unsafe.Pointer {
	return atomic.LoadPointer(&lvl[forLevel])
}

// AddPtrUnsafe sets the pointer to the next element at the
// given level. It is not safe for concurrent use.
func (lvl *Tower) AddPtrUnsafe(toLevel int, p unsafe.Pointer) {
	lvl[toLevel] = p
}

// AddPtr sets the pointer to the next element at the given level.
func (lvl *Tower) AddPtr(toLevel int, p unsafe.Pointer) {
	atomic.StorePointer(&lvl[toLevel], p)
}

// CompareAndSwapNext performs the atomic CAS operation for the element
// pointed to at the given level.
func (lvl *Tower) CompareAndSwapNext(level int, old, new unsafe.Pointer) (swapped bool) {
	return atomic.CompareAndSwapPointer(&lvl[level], old, new)
}

// UpdateNext atomically stores the given pointer at the given level.
func (lvl *Tower) UpdateNext(level int, new unsafe.Pointer) {
	atomic.StorePointer(&lvl[level], new)
}
