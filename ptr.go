package ash

import (
	"unsafe"
)

// TagPointer applies the flag to the higher 8 bits of the pointer address
//
// Example:
//
//	normal ptr ->
//	00000000_00000000_00000001_01000000_00000000_00000000_11100001_10111000
//	(ptr|2<<56) ->
//	00000010_00000000_00000001_01000000_00000000_00000000_11100001_10111000
//
//go:nocheckptr
func TagPointer(ptr unsafe.Pointer, flag int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(uintptr(ptr) | uintptr(flag)<<56))
}

// PointerFromTagPointer clears the higher 8 bits from the pointer address, effectively
// removing all of the tagging that might have been applied to it. This is useful because on
// platforms where TBI is not or cannot be enabled, calling an object using the 'tainted'
// pointer will cause a crash.
//
//go:nocheckptr
func PointerFromTagPointer(tp unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(uintptr(tp) ^ (uintptr(tp) & (0xff << 56))))
}

func IsPointerMarked(ptr unsafe.Pointer) bool {
	return ((uintptr(ptr) >> 56) & marked) != 0
}
