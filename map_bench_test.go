package ash

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkSyncMap_70Load20Store10Delete(b *testing.B) {
	var cache sync.Map
	keys := generateIntKeys(1000000)
	total := len(keys) - 1

	var (
		storeCnt  int64
		loadCnt   int64
		deleteCnt int64
		totalCnt  int64
	)

	// for i := 0; i < len(keys); i++ {
	// 	cache.Store(keys[i], i)
	// }

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var cnt int
		for pb.Next() {
			atomic.AddInt64(&totalCnt, 1)
			cnt++
			r := rand.Intn(10)
			if r < 1 {
				cache.Delete(keys[cnt%total])
				atomic.AddInt64(&deleteCnt, 1)
				continue
			}
			if r < 3 {
				cache.Store(keys[cnt%total], cnt)
				atomic.AddInt64(&storeCnt, 1)
				continue
			}
			cache.Load(keys[cnt%total])
			atomic.AddInt64(&loadCnt, 1)
		}
	})
	b.Cleanup(func() {
		b.Log("sync.Map total calls to Store/Delete/Load: ",
			atomic.LoadInt64(&storeCnt), "/",
			atomic.LoadInt64(&deleteCnt), "/",
			atomic.LoadInt64(&loadCnt), "/",
		)
		b.Log("Execution time: ", b.Elapsed())
	})
}

func BenchmarkAshMap_70Load20Store10Delete(b *testing.B) {
	cache := new(Map).From(NewSkipList(32))
	keys := generateIntKeys(1000000)
	total := len(keys) - 1

	var (
		storeCnt  int64
		loadCnt   int64
		deleteCnt int64
		totalCnt  int64
	)
	// for i := 0; i < len(keys); i++ {
	// 	cache.Store(keys[i], i)
	// }

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var cnt int
		for pb.Next() {
			atomic.AddInt64(&totalCnt, 1)
			cnt++
			r := rand.Intn(10)
			if r < 1 {
				cache.Delete(keys[cnt%total])
				atomic.AddInt64(&deleteCnt, 1)
				continue
			}
			if r < 3 {
				cache.Store(keys[cnt%total], cnt)
				atomic.AddInt64(&storeCnt, 1)
				continue
			}
			cache.Load(keys[cnt%total])
			atomic.AddInt64(&loadCnt, 1)
		}
	})
	b.Cleanup(func() {
		b.Log("ash.Map total calls to Store/Delete/Load: ",
			atomic.LoadInt64(&storeCnt), "/",
			atomic.LoadInt64(&deleteCnt), "/",
			atomic.LoadInt64(&loadCnt), "/",
		)
		//" total: ", atomic.LoadInt64(&totalCnt))
		b.Log("Execution time: ", b.Elapsed())
	})
}
