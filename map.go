package ash

import (
	"sync/atomic"
)

// Map is a drop-in replacement for sync.Map backed by skip list.
// Use &ash.Map{*SkipList} or new(ash.Map).From(ash.NewSkipList(maxlevel)) to instantiate.
type Map struct {
	*SkipList
}

func (m *Map) From(sl *SkipList) *Map {
	m.SkipList = sl
	return m
}

// Clear deletes all the entries, resulting in an empty Map.
func (m *Map) Clear() {
	for level := int(atomic.LoadUint64(&m.SkipList.topLevel) - 1); level >= 0; level-- {
		m.SkipList.start.tower.UpdateNext(level, nil)
	}
}

// CompareAndDelete deletes the entry for key if its value is equal to old.
// The old value must be of a comparable type.
//
// If there is no current value for key in the map, CompareAndDelete
// returns false (even if the old value is the nil interface value).
func (m *Map) CompareAndDelete(key, old any) (deleted bool) {
	keyHash := HashFunc(key)
	nd := m.SkipList.Search(keyHash)
	if nd == nil {
		return
	}
	val := nd.GetVal()
	if val == nil || val != old {
		return
	}
	m.Delete(key)
	deleted = true
	return
}

// CompareAndSwap swaps the old and new values for key
// if the value stored in the map is equal to old.
// The old value must be of a comparable type.
func (m *Map) CompareAndSwap(key, old, new any) (swapped bool) {
	hKey := HashFunc(key)
	for {
		nd := m.SkipList.Search(hKey)
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
func (m *Map) Delete(key any) {
	m.SkipList.Delete(HashFunc(key))
}

// Load returns the value stored in the map for a key, or nil if no value is present.
// The ok result indicates whether value was found in the map.
func (m *Map) Load(key any) (value any, ok bool) {
	nd := m.SkipList.Search(HashFunc(key))
	if nd == nil {
		return nil, false
	}
	val := nd.GetVal()
	return val, ok
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *Map) LoadAndDelete(key any) (value any, loaded bool) {
	nd, _ := m.SkipList.Delete(HashFunc(key))
	if nd == nil {
		return
	}
	return nd.GetVal(), true
}

// LoadOrStore returns the existing value for the key if present. Otherwise, it stores and returns
// the given value. The loaded result is true if the given value was loaded, false if stored.
func (m *Map) LoadOrStore(key, value any) (actual any, loaded bool) {
	hKey := HashFunc(key)
	nd := m.SkipList.Search(hKey)
	if nd != nil {
		actual = nd.GetVal()
		loaded = true
		return
	}
	m.SkipList.Store(hKey, value)
	actual = value
	return
}

// Store sets the value for a key.
func (m *Map) Store(key, value any) {
	hKey := HashFunc(key)
	m.SkipList.Store(hKey, value)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *Map) Range(f func(key, value any) bool) {
	// iterate through items at the bottom level
	for next := m.SkipList.start.Next(0); next != nil; next = next.Next(0) {
		if !f(next.key, next.GetVal()) {
			return
		}
	}
}

// Len reports the number of stored keys
func (m *Map) Len() (l int) {
	next := m.SkipList.start.Next(0)
	for next != nil {
		l++
		next = next.Next(0)
	}
	return
}

// Swap returns the value for a key and returns the previous value if any.
// The loaded result reports whether the key was present.
func (m *Map) Swap(key, value any) (previous any, loaded bool) {
	hKey := HashFunc(key)
	for {
		nd := m.SkipList.Search(hKey)
		if nd == nil {
			m.SkipList.Store(hKey, value)
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
