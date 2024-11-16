package ash

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	val    unsafe.Pointer
	levels layers
	key    uint64
	flags  uint32
}

func (nd *node) getVal() (val any) {
	return *(*any)(atomic.LoadPointer(&nd.val))
}

func (nd *node) updateVal(val any) bool {
	old := atomic.LoadPointer(&nd.val)
	return atomic.CompareAndSwapPointer(&nd.val, old, unsafe.Pointer(&val))
}

func (nd *node) swapVal(old, val any) bool {
	return atomic.CompareAndSwapPointer(&nd.val, unsafe.Pointer(&old), unsafe.Pointer(&val))
}

func (nd *node) next(fromLevel int) (n *node) {
	n = (*node)(nd.levels.next(fromLevel))
	if n == nil {
		return
	}
	for isPointerMarked(unsafe.Pointer(n)) && nd != nil {
		// lazily unlink node from the list at the current level
		nd.levels.swapNext(fromLevel, unsafe.Pointer(n), unsafe.Pointer(n.next(fromLevel)))
		n = n.next(fromLevel)
	}
	return
}

func (nd *node) add(toLevel int, next *node) {
	nd.levels.add(toLevel, unsafe.Pointer(next))
}

/*func (nd *node) isFullyLinked() bool {
	return isFlagSet(&nd.flags, fullyLinked)
}

func (nd *node) isMarked() bool {
	return isFlagSet(&nd.flags, marked)
}

func (nd *node) setFlags(f uint32) {
	setFlags(&nd.flags, f)
}

func (nd *node) unsetFlags(f uint32) {
	unsetFlags(&nd.flags, f)
}*/
