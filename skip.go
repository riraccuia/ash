package ash

import (
	"sync/atomic"
	"unsafe"
)

func (sl *Map) search(key uint64) (nd *Node) {
	nd, _, _ = sl.findNode(key, false)
	return
}

func (sl *Map) findNode(key uint64, full bool) (nd *Node, preds, succs []*Node) {
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

func (sl *Map) storeSafe(key uint64, val any) {
	var (
		nd           *Node
		preds, succs []*Node
		untilLevel   int
	)

	// start from the bottom level
	for {
		nd, preds, succs = sl.findNode(key, true)
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

		untilLevel = randomHeight(int(sl.maxLevel))
		sl.updateTopLevel(uint64(untilLevel))

		// link all the successors to the new node
		for level := 0; level < untilLevel; level++ {
			if level < len(succs) {
				nd.Add(level, succs[level])
				continue
			}
			nd.Add(level, nil)
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
			nd, preds, succs = sl.findNode(key, true)
			continue
		}
		break
	}
}

func (sl *Map) delete(key uint64) bool {
	nd, preds, _ := sl.findNode(key, true)
	for nd != nil {
		if preds[0].Next(0) != nd {
			return true
		}
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
			nd, preds, _ = sl.findNode(key, true)
			continue
		}
		// lazily attempt to unlink the bottom node
		// preds[0].tower.SwapNext(0, unsafe.Pointer(taggedPtr), unsafe.Pointer(succs[0]))
		break
	}
	return true
}

func (sl *Map) updateTopLevel(level uint64) {
	if level == sl.maxLevel {
		return
	}
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
