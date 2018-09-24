package pf

import (
	"unsafe"
)

const MAXPATHLEN = 1024

type Action uint32

const (
	PF_CHANGE_ADD_TAIL   Action = 2
	PF_CHANGE_REMOVE     Action = 5
	PF_CHANGE_GET_TICKET Action = 6
)

type RuleAction byte

const (
	PF_RDR  RuleAction = 8
	PF_PASS RuleAction = 0
)

const Pfioc_ruleSize = 3104

type Pfioc_rule struct {
	Action      uint32 // type Action
	Ticket      uint32
	Pool_ticket uint32
	Nr          uint32
	Anchor      [MAXPATHLEN]byte
	Anchor_call [MAXPATHLEN]byte
	// ^-- 2064 bytes

	Padding__ [1040]byte // struct pf_rule rule;
}

func (rule *Pfioc_rule) SetPoolTicket(pooladdr Pfioc_pooladdr) {
	rule.Pool_ticket = pooladdr.Ticket
}

func (rule *Pfioc_rule) SetAction(action Action) {
	rule.Action = uint32(action)
}

func (rule *Pfioc_rule) SetRuleAction(ruleAction RuleAction) {
	// https://github.com/apple/darwin-xnu/blob/0a798f6738bc1db01281fc08ae024145e84df927/bsd/net/pfvar.h#L749
	rule.Padding__[1004] = byte(ruleAction)
}

func (rule *Pfioc_rule) SetAnchorCall(name string) {
	if len(name) > len(rule.Anchor_call)-1 {
		panic("AnchorCall name too long")
	}
	for i, b := range []byte(name) {
		rule.Anchor_call[i] = b
	}
}

func (rule *Pfioc_rule) SetAnchor(name string) {
	if len(name) > len(rule.Anchor)-1 {
		panic("Anchor name too long")
	}
	for i, b := range []byte(name) {
		rule.Anchor[i] = b
	}
}

func init() {
	var rule Pfioc_rule
	if unsafe.Sizeof(rule) != Pfioc_ruleSize {
		panic("Expected Pfioc_rule to match")
	}
}
