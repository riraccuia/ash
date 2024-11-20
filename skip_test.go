package ash

import (
	"fmt"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

func TestSkipList_store(t *testing.T) {
	sl := NewSkipList(16)
	sl.storeSafe(3, 9876)
	sl.storeSafe(5, 9877)
	sl.storeSafe(345, 9878)
	sl.storeSafe(77, 9879)
	sl.storeSafe(342, 9880)
	t.Log("Ranging")
	sl.Range(func(key, value any) bool {
		t.Log(value.(int))
		return true
	})
	t.Logf("Length: %v", sl.Len())
	nd := sl.search(77)
	if nd == nil {
		t.Fatalf("should be found")
	}
	t.Log(nd.getVal().(int))
	sl.delete(77)
	sl.delete(77)
	nd = sl.search(77)
	if nd != nil {
		t.Fatalf("should not be found")
	}
	nd = sl.search(77)
	if nd != nil {
		t.Fatalf("should not be found")
	}
	t.Logf("Length: %v", sl.Len())
	sl.Clear()
	if sl.Len() > 0 {
		t.Fatalf("unexpected count, found: %v", sl.Len())
	}
}

func TestGetSetBit(t *testing.T) {
	var flags uint32

	flags |= fullyLinked | marked

	t.Logf("%08b", flags)

	unsetFlags(&flags, marked)

	t.Logf("%08b", flags)

	t.Logf("Marked: %v", isFlagSet(&flags, marked))

	flags ^= fullyLinked

	t.Logf("%08b", flags)

	flags ^= marked

	t.Logf("%08b", flags)

	t.Logf("Marked: %v", isFlagSet(&flags, marked))
}

func TestPointer(t *testing.T) {
	var a uint64 = 1234
	b := &a
	/*var c uint64
	  c = math.MaxUint64
	  t.Logf("%064b", c)
	  t.Logf("%064b - %b", uintptr(unsafe.Pointer(b)), uint(uintptr(unsafe.Pointer(b))))
	  t.Logf("%064b", ^uint(7))
	  t.Logf("%v vs %v", unsafe.Sizeof(c), unsafe.Sizeof(b))*/

	// **********

	/*c := GetTaggablePointer(unsafe.Pointer(b))
	  TagPointer(&c, marked)
	  t.Logf("%064b", ReturnTag(&c))
	  rp := (*uint64)(RestorePointer(&c))
	  t.Logf("%064b", rp)
	  t.Logf("%064b", c)
	  t.Log(*rp)*/

	// **********

	c := uint64(uintptr(unsafe.Pointer(b)))
	t.Logf("%064b", b)
	t.Logf("%064b", c|markTest)
	d := (*uint64)(unsafe.Pointer(uintptr(c | markTest)))
	t.Logf("%064b", d)
	t.Log(*d)
	d = (*uint64)(unsafe.Pointer(uintptr(c | marked<<56)))
	t.Logf("%064b", d)
	t.Log(*d)
	d = (*uint64)(TagPointer(unsafe.Pointer(b), marked))
	t.Logf("%064b", d)
	t.Log(*d)
	t.Logf("Marked: %v", isPointerMarked(unsafe.Pointer(d)))

	var arr []*uint64 = []*uint64{&a, (*uint64)(TagPointer(unsafe.Pointer(&a), marked))}

	for i, v := range arr {
		t.Logf("Pointer marked for [%v]: %v", i, isPointerMarked(unsafe.Pointer(v)))
	}

	t.Logf("%064b", marked<<56)
}

const markTest uint64 = 0x8000000000000000

func TestRange(t *testing.T) {
	sl := &SkipList{
		start: &node{
			key: 0,
		},
		topLevel: 3,
	}
	for i := 0; i < 3; i++ {
		Bar(sl)
		time.Sleep(time.Second)
		runtime.GC()
	}
	time.Sleep(20 * time.Second)
}

type tst struct {
	n int
}

func Bar(sl *SkipList) {
	a := &tst{
		n: 900,
	}
	runtime.SetFinalizer(a, func(f *tst) {
		fmt.Println("freeing memory")
	})
	sl.storeSafe(uint64(RandInt()), a)
	sl.Clear()
	a = nil
}
