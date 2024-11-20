package ash

import (
	//"errors"

	"math"
	"sync/atomic"
	"unsafe"
)

const (
	PValue = 0.5 // p = 1/2
)

var probabilities [64]uint32

func init() {
	probability := 1.0

	for level := 0; level < 64; level++ {
		probabilities[level] = uint32(probability * float64(math.MaxUint32))
		probability *= PValue
	}
}

type SkipList struct {
	start    *node
	topLevel uint64
	maxLevel uint64
}

func NewSkipList(maxLevel int) *SkipList {
	if maxLevel < 1 || maxLevel > 64 {
		panic("maxLevel must be between 1 and 64")
	}
	sl := &SkipList{
		topLevel: 1,
		maxLevel: uint64(maxLevel),
	}
	sl.start = &node{
		key:   0,
		tower: NewTower(),
	}
	for i := 0; i < maxLevel; i++ {
		sl.start.tower.add(i, nil)
	}
	return sl
}

func randomHeight(maxLevel int) int {
	seed := fastrand()

	height := 1
	for height < maxLevel && seed <= probabilities[height] {
		height++
	}

	return height
}

func (sl *SkipList) updateTopLevel(level uint64) {
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

func (sl *SkipList) Clear() {
	for level := int(atomic.LoadUint64(&sl.topLevel) - 1); level >= 0; level-- {
		sl.start.tower.updateNext(level, nil)
	}
}

// CompareAndDelete deletes the entry for key if its value is equal to old.
// The old value must be of a comparable type.
//
// If there is no current value for key in the map, CompareAndDelete
// returns false (even if the old value is the nil interface value).
func (sl *SkipList) CompareAndDelete(key, old any) (deleted bool) {
	keyHash := HashFunc(key)
	nd := sl.search(keyHash)
	if nd == nil {
		return
	}
	ok := nd.swapVal(old, nil)
	if !ok {
		return
	}
	sl.Delete(key)
	deleted = true
	return
}

// CompareAndSwap swaps the old and new values for key
// if the value stored in the map is equal to old.
// The old value must be of a comparable type.
func (sl *SkipList) CompareAndSwap(key, old, new any) (swapped bool) {
	keyHash := HashFunc(key)
	nd := sl.search(keyHash)
	if nd == nil {
		return
	}
	return nd.swapVal(old, new)
}

func (sl *SkipList) Load(key any) (value any, ok bool) {
	keyHash := HashFunc(key)
	nd := sl.search(keyHash)
	if nd == nil {
		return nil, false
	}
	val := nd.getVal()
	return val, ok
}

func (sl *SkipList) Store(key any, item interface{}, lifetime int64) {
	hKey := HashFunc(key)
	sl.storeSafe(hKey, item)
}

func (sl *SkipList) Range(f func(key, value any) bool) {
	// iterate through items at the base level
	for next := sl.start.next(0); next != nil; next = next.next(0) {
		if !f(next.key, next.getVal()) {
			return
		}
	}
}

func (sl *SkipList) Delete(key any) {
	sl.delete(HashFunc(key))
}

func (sl *SkipList) Len() (l int) {
	next := sl.start.next(0)
	for next != nil {
		l++
		next = next.next(0)
	}
	return
}

func (sl *SkipList) search(key uint64) (nd *node) {
	nd, _, _ = sl.findNode(key, false)
	return
}

func (sl *SkipList) findNode(key uint64, full bool) (nd *node, pred, succ []*node) {
	var (
		lev      *level
		prev     = sl.start
		topLevel = int(atomic.LoadUint64(&sl.topLevel))
	)
	pred, succ = make([]*node, topLevel), make([]*node, topLevel)
	for lev = prev.tower.getLevel(topLevel - 1); lev != nil; lev = lev.down() {
		next := (*node)(lev.next())
		for next != nil && key > next.key {
			prev = next
			lev = prev.tower.top
			next = (*node)(lev.next())
		}
		pred[lev.id] = prev
		succ[lev.id] = next
		if next != nil && key == next.key {
			nd = next
			succ[lev.id] = (*node)(next.tower.top.next())
			if !full {
				return
			}
		}
	}
	return
}

func (sl *SkipList) storeSafe(key uint64, val interface{}) {
	var (
		nd           *node
		preds, succs []*node
		untilLevel   int
	)

	// start from the bottom level
	for {
		nd, preds, succs = sl.findNode(key, true)
		if nd != nil {
			if !nd.updateVal(val) {
				// cas failed, start over
				continue
			}
			return
		}
		nd = &node{
			key:   key,
			val:   unsafe.Pointer(&val),
			tower: NewTower(),
		}

		untilLevel = randomHeight(int(sl.maxLevel))
		sl.updateTopLevel(uint64(untilLevel))

		// link all the successors to the new node
		for level := 0; level < untilLevel; level++ {
			var next *node
			if level < len(succs) {
				next = succs[level]
			}
			nd.add(level, next)
		}
		// atomically link the node to its predecessor
		if !preds[0].tower.swapNext(0, unsafe.Pointer(succs[0]), unsafe.Pointer(nd)) {
			// cas failed, need to start over
			continue
		}
		// made it: the node now physically exists in the skip list
		break
	}

	// update the predecessors in the upper levels
	for nd != nil {
		var (
			retry bool
			until = len(preds) % untilLevel
		)
		for level := 1; level < until; level++ {
			pred := preds[level]
			succ := succs[level]
			if pred == nil {
				pred = sl.start
			}
			if !pred.tower.swapNext(level, unsafe.Pointer(succ), unsafe.Pointer(nd)) {
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

func (sl *SkipList) delete(key uint64) bool {
	nd, prevs, succs := sl.findNode(key, true)
	for nd != nil {
		if prevs[0].next(0) != nd {
			return true
		}
		var (
			retry bool
			// obtain a marked pointer to the node, it has the same memory address
			taggedPtr = (*node)(TagPointer(unsafe.Pointer(nd), marked))
		)
		for level := len(prevs) - 1; level > 0; level-- {
			prev := prevs[level]
			if prev == nil {
				continue
			}
			if prev.next(level) != nd {
				continue
			}
			// log.Printf("%064b\n", unsafe.Pointer(nd))
			if !prev.tower.swapNext(level, unsafe.Pointer(nd), unsafe.Pointer(taggedPtr)) {
				retry = true
				/*fmt.Println("can't mark level", level)
				  log.Printf("%064b\n", unsafe.Pointer(prev.next(level)))
				  log.Printf("%064b\n", unsafe.Pointer(nd))
				  log.Printf("%064b\n", unsafe.Pointer(taggedPtr))
				  log.Printf("%v\n", prev.next(level).getVal().(int))
				  log.Printf("%v\n", nd.getVal().(int))*/
				break
			}
			// log.Printf("%064b\n", prev.tower.next(level))
			// lazily attempt to unlink the node
			prev.tower.swapNext(level, unsafe.Pointer(taggedPtr), unsafe.Pointer(succs[level]))
		}
		if retry || !prevs[0].tower.swapNext(0, unsafe.Pointer(nd), unsafe.Pointer(taggedPtr)) {
			nd, prevs, succs = sl.findNode(key, true)
			continue
		}
		break
	}
	nd = nil
	return true
}
