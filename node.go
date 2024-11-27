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
		next  = (*Node)(lev.NextPtr())
		_next = next
	)
	for next != nil && IsPointerMarked(unsafe.Pointer(next)) {
		next = (*Node)(next.tower.NextPtr(lev.id))
	}
	if next != _next {
		// log.Println("unlinking node")
		// lazily unlink node from the list at the current level
		nd.tower.SwapNext(lev.id, unsafe.Pointer(_next), unsafe.Pointer(next))
	}
	return next
}

func (nd *Node) Next(forLevel int) *Node {
	var (
		next  = (*Node)(nd.tower.NextPtr(forLevel))
		_next = next
	)
	for next != nil && IsPointerMarked(unsafe.Pointer(next)) {
		next = next.Next(forLevel)
	}
	if next != _next {
		// lazily unlink node from the list at the current level
		nd.tower.SwapNext(forLevel, unsafe.Pointer(next), unsafe.Pointer(_next))
	}
	return next
}

func (nd *Node) AddNext(toLevel int, next *Node) {
	nd.tower.AddPtr(toLevel, unsafe.Pointer(next))
}
