package ash

import (
	_ "unsafe"
)

//go:noescape
//go:linkname fastrand runtime.fastrand
func fastrand() uint32
