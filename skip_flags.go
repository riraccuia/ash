package ash

import "sync/atomic"

const (
	fullyLinked = 1 << iota // 0b001
	marked                  // 0b010
)

func isFlagSet(b *uint32, flags uint32) bool {
	return (atomic.LoadUint32(b) & flags) != 0
}

func setFlags(b *uint32, flags uint32) {
	for {
		v := atomic.LoadUint32(b)
		newV := v | flags
		if atomic.CompareAndSwapUint32(b, v, newV) {
			break
		}
	}
}

func unsetFlags(b *uint32, flags uint32) {
	for {
		v := atomic.LoadUint32(b)
		if (v & flags) == 0 {
			break
		}
		newV := v ^ (v & flags)
		if atomic.CompareAndSwapUint32(b, v, newV) {
			break
		}
	}
}
