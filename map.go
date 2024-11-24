package ash

import (
	"sync/atomic"
)

type Map struct {
	start    *Node
	topLevel uint64
	maxLevel uint64
}

func NewMap(maxLevel int) *Map {
	if maxLevel < 1 || maxLevel > 64 {
		panic("maxLevel must be between 1 and 64")
	}
	sl := &Map{
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

// Clear deletes all the entries, resulting in an empty Map.
func (sl *Map) Clear() {
	for level := int(atomic.LoadUint64(&sl.topLevel) - 1); level >= 0; level-- {
		sl.start.tower.UpdateNext(level, nil)
	}
}

// CompareAndDelete deletes the entry for key if its value is equal to old.
// The old value must be of a comparable type.
//
// If there is no current value for key in the map, CompareAndDelete
// returns false (even if the old value is the nil interface value).
func (sl *Map) CompareAndDelete(key, old any) (deleted bool) {
	keyHash := HashFunc(key)
	nd := sl.search(keyHash)
	if nd == nil {
		return
	}
	ok := nd.SwapVal(old, nil)
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
func (sl *Map) CompareAndSwap(key, old, new any) (swapped bool) {
	keyHash := HashFunc(key)
	nd := sl.search(keyHash)
	if nd == nil {
		return
	}
	return nd.SwapVal(old, new)
}

// Load returns the value stored in the map for a key, or nil if no value is present.
// The ok result indicates whether value was found in the map.
func (sl *Map) Load(key any) (value any, ok bool) {
	keyHash := HashFunc(key)
	nd := sl.search(keyHash)
	if nd == nil {
		return nil, false
	}
	val := nd.GetVal()
	return val, ok
}

// Store sets the value for a key.
func (sl *Map) Store(key, value any) {
	hKey := HashFunc(key)
	sl.storeSafe(hKey, value)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (sl *Map) Range(f func(key, value any) bool) {
	// iterate through items at the bottom level
	for next := sl.start.Next(0); next != nil; next = next.Next(0) {
		if !f(next.key, next.GetVal()) {
			return
		}
	}
}

// Delete deletes the value for a key.
func (sl *Map) Delete(key any) {
	sl.delete(HashFunc(key))
}

// Len reports the number of stored keys
func (sl *Map) Len() (l int) {
	next := sl.start.Next(0)
	for next != nil {
		l++
		next = next.Next(0)
	}
	return
}
