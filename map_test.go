package ash

import (
	"strconv"
	"sync"
	"testing"
)

func TestMapStore(t *testing.T) {
	m := new(Map).From(NewSkipList(32))
	keys := generateKeys(1000000)

	wg := sync.WaitGroup{}

	for j := 0; j < len(keys); j++ {
		wg.Add(1)
		x := j
		go func() {
			m.Store(keys[x], keys[x])
			wg.Done()
		}()
	}

	wg.Wait()

	for _, v := range keys {
		r, _ := m.Load(v)
		if r.(string) != v {
			t.Errorf("unexpected value for key %v = %v\n", v, r)
		}
	}
}

func generateKeys(numKeys int) []string {
	var keys []string
	for i := 0; i < numKeys; i++ {
		keys = append(keys, ("key-" + strconv.Itoa(i)))
	}
	return keys
}

func generateIntKeys(numKeys int) []int {
	var keys []int
	for i := 0; i < numKeys; i++ {
		keys = append(keys, i)
	}
	return keys
}
