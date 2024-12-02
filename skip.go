package ash

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

// SkipList is a lock-free, concurrent safe skip list implementation.
type SkipList struct {
	start    Node
	topLevel uint64
	maxLevel uint64
}

// NewSkipList returns a lock-free, concurrent safe skip list with its maxlevel set to the given height.
// For p = 1/2, using maxlevel=16 should be appropriate for storing up to 2^16 elements.
// The p value is controlled by the `PValue` global and defaults to 1/2. Maxlevel is a number between 1 and 64.
func NewSkipList(maxLevel int) *SkipList {
	if maxLevel < 1 || maxLevel > CapLevel {
		panic("maxLevel must be between 1 and 64")
	}
	sl := &SkipList{
		topLevel: 1,
		maxLevel: uint64(maxLevel),
	}
	sl.start = Node{
		key: 0,
	}
	return sl
}

// Search calls FindNode() with the 'full' parameter set to false
// Returns the node, if found, or nil.
func (sl *SkipList) Search(key uint64) (nd *Node) {
	nd = sl.FindNode(key, false, nil, nil)
	return
}

// FindNode performs a search from the top level to the bottom level to find the target element.
// Nodes with a reference mark will be ignored. The 'full' parameter, when set to false, makes
// FindNode stop and return when the element is found, without reaching to the bottom of the list.
// If the element is found, it is returned along with its predecessors and successors on all levels,
// note that on levels where the element is found, the successor is the node (element) itself.
// The preds and succs arguments are allocated by the caller, or set to nil.
func (sl *SkipList) FindNode(key uint64, full bool, preds, succs *[CapLevel]*Node) (nd *Node) {
	p := &sl.start
	for level := int(atomic.LoadUint64(&sl.topLevel)) - 1; level >= 0; level-- {
		next := p.Next(level)
		for next != nil && key > next.key {
			p = next
			next = p.Next(level)
		}

		if preds != nil && succs != nil {
			preds[level] = p
			succs[level] = next
		}

		if next == nil || key != next.key {
			continue
		}

		nd = next

		if !full {
			break
		}
	}
	return
}

// Store adds the element to the list. If the element exists, its value is updated.
// Otherwise it finds its predecessors and successors in all levels, creates the new node
// with all pointers pointing to the potential successors. It will first try to add the
// new node using CAS to the bottom level, then it will modify the pointers in all the
// previous nodes to point to the current node.
func (sl *SkipList) Store(key uint64, val any) {
	var (
		nd           *Node
		preds, succs [CapLevel]*Node
		untilLevel   int
	)

	// start from the bottom level
	for {
		nd = sl.FindNode(key, true, &preds, &succs)
		if nd != nil {
			if !nd.UpdateVal(val) {
				// cas failed, start over
				continue
			}
			return
		}

		nd = &Node{
			key: key,
			val: unsafe.Pointer(&val),
		}

		untilLevel = RandomHeightFunc(int(sl.maxLevel))
		sl.updateTopLevel(uint64(untilLevel))

		// link all the successors to the new node
		for level := 0; level < untilLevel; level++ {
			nd.tower.AddPtrUnsafe(level, unsafe.Pointer(succs[level]))
		}
		// atomically link the node to its predecessor in the bottom level
		if !preds[0].tower.SwapNext(0, unsafe.Pointer(succs[0]), unsafe.Pointer(nd)) {
			// cas failed, need to start over
			continue
		}
		// made it: the node now physically exists in the skip list
		break
	}

	// update the predecessors in the upper levels starting from the top
	for nd != nil {
		var retry bool
		for level := 1; level < untilLevel; level++ {
			pred, succ := preds[level], succs[level]
			if pred == nil {
				pred = &sl.start
			}
			if !pred.tower.SwapNext(level, unsafe.Pointer(succ), unsafe.Pointer(nd)) {
				retry = true
				break
			}
		}
		if retry {
			// find up to date predecessors
			nd = sl.FindNode(key, true, &preds, &succs)
			continue
		}
		break
	}
}

// Delete removes the element from the list.
// It will start to mark from the top level until one level above the bottom level.
// After all upper level references are marked, it marks the bottom level to indicate
// logical deletion from the list.
func (sl *SkipList) Delete(key uint64) (nd *Node, deleted bool) {
	var preds, succs [CapLevel]*Node
	nd = sl.FindNode(key, true, &preds, &succs)
	for nd != nil {
		var (
			retry bool
			// obtain a marked pointer to the node, it has the same memory address
			taggedPtr = TagPointer(unsafe.Pointer(nd), marked)
		)

		for level := int(atomic.LoadUint64(&sl.topLevel)) - 1; level >= 0; level-- {
			prev, succ := preds[level], succs[level]
			if prev == nil || succ != nd {
				continue
			}
			if !prev.tower.SwapNext(level, unsafe.Pointer(nd), taggedPtr) {
				retry = true
				break
			}
			nxt := unsafe.Pointer(nd.Next(level))
			// now that we marked the node, it is safe to unlink it
			if !prev.tower.SwapNext(level, taggedPtr, nxt) {
				retry = true
				break
			}
		}
		if retry {
			runtime.Gosched()
			nd = sl.FindNode(key, true, &preds, &succs)
			continue
		}
		deleted = true
		break
	}
	return
}

func (sl *SkipList) updateTopLevel(level uint64) {
	switch {
	case level > sl.maxLevel:
		panic("unexpected height")
	default:
		for {
			topLevel := atomic.LoadUint64(&sl.topLevel)
			if level <= topLevel {
				break
			}
			if atomic.CompareAndSwapUint64(&sl.topLevel, topLevel, level) {
				break
			}
		}
	}
}
