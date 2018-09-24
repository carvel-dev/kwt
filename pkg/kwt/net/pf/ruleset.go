package pf

import (
	"unsafe"
)

const (
	PF_ANCHOR_NAME_SIZE = 64
	Pfioc_rulesetSize   = 1092
)

type Pfioc_ruleset struct {
	Nr   uint32
	Path [MAXPATHLEN]byte
	Name [PF_ANCHOR_NAME_SIZE]byte
}

func (ruleset Pfioc_ruleset) NameString() string {
	// Construct string without NULs at the end, as otherwise
	// regular string comparison does not work due to length mismatch
	var str string
	for _, b := range ruleset.Name {
		if b != 0 {
			str += string(b)
		}
	}
	return str
}

func init() {
	var rule Pfioc_ruleset
	if unsafe.Sizeof(rule) != Pfioc_rulesetSize {
		panic("Expected Pfioc_ruleset to match")
	}
}
