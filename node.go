package ash

import (
	"sync/atomic"
	"unsafe"
)

type Node struct {
	val   unsafe.Pointer
	tower *Tower
	key   uint64
}

func (nd *Node) GetVal() (val any) {
	return *(*any)(atomic.LoadPointer(&nd.val))
}

func (nd *Node) UpdateVal(val any) bool {
	old := atomic.LoadPointer(&nd.val)
	return atomic.CompareAndSwapPointer(&nd.val, old, unsafe.Pointer(&val))
}

func (nd *Node) SwapVal(old, val any) bool {
	return atomic.CompareAndSwapPointer(&nd.val, unsafe.Pointer(&old), unsafe.Pointer(&val))
}

func (nd *Node) NextFromLevel(lev *Level) *Node {
	var (
		next  = lev.NextPtr()
		_next = next
	)
	for next != nil && IsPointerMarked(next) {
		next = (*Node)(PointerFromTagPointer(next)).tower.NextPtr(lev.id)
	}
	if next != _next {
		// log.Println("unlinking node")
		// lazily unlink node from the list at the current level
		nd.tower.SwapNext(lev.id, _next, next)
	}
	return (*Node)(next)
}

func (nd *Node) Next(forLevel int) *Node {
	var (
		next  = nd.tower.NextPtr(forLevel)
		_next = next
	)
	for next != nil && IsPointerMarked(next) {
		next = (*Node)(PointerFromTagPointer(next)).tower.NextPtr(forLevel)
	}
	if next != _next {
		// lazily unlink node from the list at the current level
		nd.tower.SwapNext(forLevel, next, _next)
	}
	return (*Node)(next)
}

func (nd *Node) AddNext(toLevel int, next *Node) {
	nd.tower.AddPtr(toLevel, unsafe.Pointer(next))
}
