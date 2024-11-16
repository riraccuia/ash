package ash

import (
	"fmt"
	"unsafe"
)

// NumericToBytesUnsafe returns the underlying bytes for a numeric input
// Do not modify the return memory
func NumericToBytesUnsafe(numeric any) []byte {
	switch v := numeric.(type) {
	case int64, uint64, float64:
		return (*[8]byte)(unsafe.Pointer(&v))[:8:8]
	case int32, uint32, float32:
		return (*[4]byte)(unsafe.Pointer(&v))[:4:4]
	case int16, uint16:
		return (*[2]byte)(unsafe.Pointer(&v))[:2:2]
	case int8, uint8:
		return (*[1]byte)(unsafe.Pointer(&v))[:1:1]
	default:
		panic(fmt.Sprintf("non numeric type provided <%T>", v))
	}
}

//go:noescape
//go:linkname fastrand runtime.fastrand
func fastrand() uint32

func RandInt() int {
	x, y := fastrand(), fastrand()    // 32-bit halves
	u := uint(x)<<31 ^ uint(int32(y)) // full uint, even on 64-bit systems; avoid 32-bit shift on 32-bit systems
	i := int(u >> 1)                  // clear sign bit, even on 32-bit systems
	return i
}
