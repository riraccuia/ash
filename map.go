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
	val := nd.GetVal()
	if val == nil || val != old {
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
	hKey := HashFunc(key)
	for {
		nd := sl.search(hKey)
		if nd == nil {
			return
		}
		val := nd.GetVal()
		if val != old {
			return
		}
		if !nd.SwapVal(val, new) {
			continue
		}
		swapped = true
		break
	}
	return
}

// Delete deletes the value for a key.
func (sl *Map) Delete(key any) {
	sl.delete(HashFunc(key))
}

// Load returns the value stored in the map for a key, or nil if no value is present.
// The ok result indicates whether value was found in the map.
func (sl *Map) Load(key any) (value any, ok bool) {
	nd := sl.search(HashFunc(key))
	if nd == nil {
		return nil, false
	}
	val := nd.GetVal()
	return val, ok
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (sl *Map) LoadAndDelete(key any) (value any, loaded bool) {
	nd, _ := sl.delete(HashFunc(key))
	if nd == nil {
		return
	}
	return nd.GetVal(), true
}

// LoadOrStore returns the existing value for the key if present. Otherwise, it stores and returns
// the given value. The loaded result is true if the given value was loaded, false if stored.
func (sl *Map) LoadOrStore(key, value any) (actual any, loaded bool) {
	hKey := HashFunc(key)
	nd := sl.search(hKey)
	if nd != nil {
		actual = nd.GetVal()
		loaded = true
		return
	}
	sl.storeSafe(hKey, value)
	actual = value
	return
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

// Len reports the number of stored keys
func (sl *Map) Len() (l int) {
	next := sl.start.Next(0)
	for next != nil {
		l++
		next = next.Next(0)
	}
	return
}

// Swap returns the value for a key and returns the previous value if any.
// The loaded result reports whether the key was present.
func (sl *Map) Swap(key, value any) (previous any, loaded bool) {
	hKey := HashFunc(key)
	for {
		nd := sl.search(hKey)
		if nd == nil {
			sl.storeSafe(hKey, value)
			return
		}
		previous = nd.GetVal()
		if !nd.SwapVal(previous, value) {
			previous = nil
			continue
		}
		loaded = true
		break
	}
	return
}
