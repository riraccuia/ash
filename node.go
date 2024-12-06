package ash

import (
	"sync/atomic"
	"unsafe"
)

type Node struct {
	val   unsafe.Pointer
	tower Tower
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

func (nd *Node) Next(forLevel int) *Node {
	next := nd.tower.NextPtr(forLevel)
	// _next := next
	for next != nil && IsPointerMarked(next) {
		// log.Printf("lev %v skipping marked node (2): %#v", forLevel, (*Node)(PointerFromTagPointer(next)).GetVal())
		next = (*Node)(PointerFromTagPointer(next)).tower.NextPtr(forLevel)
	}
	// if next != _next {
	// 	nd.tower.SwapNext(forLevel, _next, next)
	// }
	return (*Node)(next)
}

func (nd *Node) AddNext(toLevel int, next *Node) {
	nd.tower.AddPtr(toLevel, unsafe.Pointer(next))
}