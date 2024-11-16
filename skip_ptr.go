package ash

import "unsafe"

// TagPointer applies the flag to the higher 8 bits of the pointer address
//
// Example:
//
//	normal ptr ->
//	00000000_00000000_00000001_01000000_00000000_00000000_11100001_10111000
//	(ptr|2<<56) ->
//	00000010_00000000_00000001_01000000_00000000_00000000_11100001_10111000
func TagPointer(ptr unsafe.Pointer, flag int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) | uintptr(flag)<<56)
}

func isPointerMarked(ptr unsafe.Pointer) bool {
	return ((uintptr(ptr) >> 56) & marked) != 0
}

/*
func GetTaggablePointer(ptr unsafe.Pointer) uint64 {
    return uint64(uintptr(ptr))
}

func RestorePointer(ptr *uint64) unsafe.Pointer {
    return unsafe.Pointer(uintptr(ReturnPointer(ptr)))
}

func TagPointer(ptr *uint64, i uint64) {
    x := *ptr | i
    *ptr = x
}

func ReturnPointer(ptr *uint64) uint64 {
    return *ptr &^ uint64(7)
}

func ReturnTag(ptr *uint64) uint64 {
    return *ptr & uint64(7)
}
*/
