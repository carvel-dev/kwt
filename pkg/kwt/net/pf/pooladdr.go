package pf

import (
	"unsafe"
)

const Pfioc_pooladdrSize = 1136

type Pfioc_pooladdr struct {
	Action   uint32
	Ticket   uint32
	Nr       uint32
	R_num    uint32
	R_action uint8
	R_last   uint8
	Af       uint8

	Padding__ [1117]byte // ...
}

func init() {
	var pooladdr Pfioc_pooladdr
	if unsafe.Sizeof(pooladdr) != Pfioc_pooladdrSize {
		panic("Expected Pfioc_pooladdrSize to match")
	}
}
