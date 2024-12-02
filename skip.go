package ash

import (
	"sync/atomic"
	"unsafe"
)

// SkipList is a lock-free, concurrent safe skip list implementation.
type SkipList struct {
	start    *Node
	topLevel uint64
	maxLevel uint64
}

// NewSkipList returns a lock-free, concurrent safe skip list with its maxlevel set to the given height.
// For p = 1/2, using maxlevel=16 should be appropriate for storing up to 2^16 elements.
// The p value is controlled by the `PValue` global and defaults to 1/2. Maxlevel is a number between 1 and 64.
func NewSkipList(maxLevel int) *SkipList {
	if maxLevel < 1 || maxLevel > 64 {
		panic("maxLevel must be between 1 and 64")
	}
	sl := &SkipList{
		topLevel: 1,
		maxLevel: uint64(maxLevel),
	}
	sl.start = &Node{
		key:   0,
		tower: NewTower(),
	}
	for i := 0; i < maxLevel; i++ {
		sl.start.tower.AddPtr(i, nil)
	}
	return sl
}

// Search calls FindNode() with the 'full' parameter set to false
// Returns the node, if found, or nil.
func (sl *SkipList) Search(key uint64) (nd *Node) {
	nd, _, _ = sl.FindNode(key, false)
	return
}

// FindNode performs a search from the top level to the bottom level to find the target element.
// Nodes with a reference mark will be ignored. The 'full' parameter, when set to false, makes
// FindNode stop and return when the element is found, without reaching to the bottom of the list.
func (sl *SkipList) FindNode(key uint64, full bool) (nd *Node, preds, succs []*Node) {
	var (
		lev      *Level
		prev     = sl.start
		topLevel = int(atomic.LoadUint64(&sl.topLevel))
	)

	preds, succs = make([]*Node, topLevel), make([]*Node, topLevel)

	for lev = prev.tower.GetLevel(topLevel - 1); lev != nil; lev = lev.Down() {
		next := prev.NextFromLevel(lev)

		// walk through the list at the current level
		for next != nil && key > next.key {
			prev = next
			lev = prev.tower.top
			next = prev.NextFromLevel(lev)
		}

		preds[lev.id], succs[lev.id] = prev, next

		if next == nil || key != next.key {
			continue
		}

		// log.Println("found full=", full, "lev", lev.id)
		nd = next

		if full {
			succs[lev.id] = nd.Next(lev.id)
			continue
		}

		succs[lev.id] = nd.NextFromLevel(nd.tower.top)
		// log.Println("found lev", lev.id)
		return
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
		preds, succs []*Node
		untilLevel   int
	)

	// start from the bottom level
	for {
		nd, preds, succs = sl.FindNode(key, true)
		if nd != nil {
			if !nd.UpdateVal(val) {
				// cas failed, start over
				continue
			}
			return
		}

		nd = &Node{
			key:   key,
			val:   unsafe.Pointer(&val),
			tower: NewTower(),
		}

		untilLevel = RandomHeightFunc(int(sl.maxLevel))
		sl.updateTopLevel(uint64(untilLevel))

		// link all the successors to the new node
		for level := 0; level < untilLevel; level++ {
			if level < len(succs) {
				nd.AddNext(level, succs[level])
				continue
			}
			nd.AddNext(level, nil)
		}
		// atomically link the node to its predecessor in the bottom level
		if !preds[0].tower.SwapNext(0, unsafe.Pointer(succs[0]), unsafe.Pointer(nd)) {
			// cas failed, need to start over
			continue
		}
		// made it: the node now physically exists in the skip list
		break
	}

	// update the predecessors in the upper levels
	for nd != nil {
		var retry bool
		for level := 1; level < untilLevel; level++ {
			var pred, succ *Node
			if level < len(succs) {
				pred, succ = preds[level], succs[level]
			}
			if pred == nil {
				pred = sl.start
			}
			if !pred.tower.SwapNext(level, unsafe.Pointer(succ), unsafe.Pointer(nd)) {
				retry = true
				break
			}
		}
		if retry {
			// find up to date predecessors
			nd, preds, succs = sl.FindNode(key, true)
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
	var preds []*Node
	nd, preds, _ = sl.FindNode(key, true)
	for nd != nil {
		// if preds[0].Next(0) != nd {
		// 	deleted = true
		// 	return
		// }
		var (
			retry bool
			// obtain a marked pointer to the node, it has the same memory address
			taggedPtr = (*Node)(TagPointer(unsafe.Pointer(nd), marked))
		)
		for level := len(preds) - 1; level > 0; level-- {
			prev := preds[level]
			if prev == nil || prev.Next(level) != nd {
				continue
			}
			if !prev.tower.SwapNext(level, unsafe.Pointer(nd), unsafe.Pointer(taggedPtr)) {
				retry = true
				break
			}
			// lazily attempt to unlink the node
			// prev.tower.SwapNext(level, unsafe.Pointer(taggedPtr), unsafe.Pointer(succs[level]))
		}
		if retry || !preds[0].tower.SwapNext(0, unsafe.Pointer(nd), unsafe.Pointer(taggedPtr)) {
			nd, preds, _ = sl.FindNode(key, true)
			continue
		}
		// lazily attempt to unlink the bottom node
		// preds[0].tower.SwapNext(0, unsafe.Pointer(taggedPtr), unsafe.Pointer(succs[0]))
		break
	}
	deleted = nd != nil
	return
}

func (sl *SkipList) updateTopLevel(level uint64) {
	switch {
	case level == sl.maxLevel:
		return
	case level < sl.maxLevel:
		for {
			topLevel := atomic.LoadUint64(&sl.topLevel)
			if level <= topLevel {
				break
			}
			if atomic.CompareAndSwapUint64(&sl.topLevel, topLevel, level) {
				break
			}
		}
	default:
		panic("unexpected height")
	}
}
