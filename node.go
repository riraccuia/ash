package ash

import (
	"sync/atomic"
	"unsafe"
)

// Node represents a node in a skip list, all of its methods.
// are concurrent safe
type Node struct {
	val   unsafe.Pointer
	tower Tower
	key   uint64
}

// GetVal retrieves the node's value atomically.
func (nd *Node) GetVal() (val any) {
	return *(*any)(atomic.LoadPointer(&nd.val))
}

// UpdateVal updates the node's value atomically, it returns
// whether the operation succeded.
func (nd *Node) UpdateVal(val any) bool {
	old := atomic.LoadPointer(&nd.val)
	return atomic.CompareAndSwapPointer(&nd.val, old, unsafe.Pointer(&val))
}

// SwapVal atomically swaps (using CAS) the node's value, it returns
// whether the operation succeded.
func (nd *Node) SwapVal(old, val any) bool {
	return atomic.CompareAndSwapPointer(&nd.val, unsafe.Pointer(&old), unsafe.Pointer(&val))
}

// Next returns the next element, at the given level, that this node
// is pointing to. It will step over marked nodes that may be found
// while walking the tree.
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

// AddNext sets the next element for this node at the given level
func (nd *Node) AddNext(toLevel int, next *Node) {
	nd.tower.AddPtr(toLevel, unsafe.Pointer(next))
}
