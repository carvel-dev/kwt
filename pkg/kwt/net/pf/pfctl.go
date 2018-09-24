package pf

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

// https://github.com/apple/darwin-xnu/blob/0a798f6738bc1db01281fc08ae024145e84df927/bsd/net/pfvar.h
// (also https://www.mirbsd.org/htman/i386/man4/pf.htm)
// (also https://www.qnx.com/developers/docs/6.4.1/neutrino/utilities/p/pf.html)
var (
	DIOCBEGINADDRS  uint32 = _IOC(IOC_INOUT, IOC_GROUP_D, 51, Pfioc_pooladdrSize)
	DIOCCHANGERULE  uint32 = _IOC(IOC_INOUT, IOC_GROUP_D, 26, Pfioc_ruleSize)
	DIOCNATLOOK     uint32 = _IOC(IOC_INOUT, IOC_GROUP_D, 23, Pfioc_natlookSize)
	DIOCGETRULESETS uint32 = _IOC(IOC_INOUT, IOC_GROUP_D, 58, Pfioc_rulesetSize)
	DIOCGETRULESET  uint32 = _IOC(IOC_INOUT, IOC_GROUP_D, 59, Pfioc_rulesetSize)
)

const (
	PF_OUT = 2
)

type Pfctl struct {
	ioctl *Ioctl
}

func NewPfctl() (*Pfctl, error) {
	ioctl, err := NewIoctl("/dev/pf")
	if err != nil {
		return nil, fmt.Errorf("opening /dev/pf: %s", err)
	}

	return &Pfctl{ioctl}, nil
}

func (p *Pfctl) Close() error {
	return p.ioctl.Close()
}

func (p *Pfctl) AddAnchorIfNotExist(name string, ruleAction RuleAction) error {
	var pooladdr Pfioc_pooladdr

	err := p.ioctl.Read(DIOCBEGINADDRS, unsafe.Pointer(&pooladdr))
	if err != nil {
		return fmt.Errorf("begin addrs: %s", err)
	}

	var rule Pfioc_rule

	rule.SetAction(PF_CHANGE_GET_TICKET)
	rule.SetPoolTicket(pooladdr)
	rule.SetAnchorCall(name)
	rule.SetRuleAction(ruleAction)

	err = p.ioctl.Read(DIOCCHANGERULE, unsafe.Pointer(&rule))
	if err != nil {
		return fmt.Errorf("change rule (get ticket): %s", err)
	}

	rule.SetAction(PF_CHANGE_ADD_TAIL)

	err = p.ioctl.Read(DIOCCHANGERULE, unsafe.Pointer(&rule))
	if err != nil {
		return fmt.Errorf("change rule (add tail): %s", err)
	}

	return nil
}

func (p *Pfctl) DeleteAnchorIfExists(name string, ruleActions []RuleAction) error {
	ruleset, err := p.findAnchorRuleset(name)
	if err != nil || ruleset == nil {
		return err
	}

	for _, ruleAction := range ruleActions {
		var rule Pfioc_rule

		rule.Nr = (*ruleset).Nr
		rule.SetAction(PF_CHANGE_GET_TICKET)
		rule.SetRuleAction(ruleAction)

		err = p.ioctl.Read(DIOCCHANGERULE, unsafe.Pointer(&rule))
		if err != nil {
			return fmt.Errorf("change rule (get ticket): %s", err)
		}

		rule.SetAction(PF_CHANGE_REMOVE)

		// TODO it appears that if there are connections in TIME_WAIT state
		// then following 'ioctl error: invalid argument' error will occur
		err = p.ioctl.Read(DIOCCHANGERULE, unsafe.Pointer(&rule))
		if err != nil {
			return fmt.Errorf("change rule (remove): %s", err)
		}
	}

	return nil
}

func (p *Pfctl) findAnchorRuleset(name string) (*Pfioc_ruleset, error) {
	var allRulesets Pfioc_ruleset

	err := p.ioctl.Read(DIOCGETRULESETS, unsafe.Pointer(&allRulesets))
	if err != nil {
		return nil, fmt.Errorf("get rulesets: %s", err)
	}

	var i uint32

	for ; i < allRulesets.Nr; i++ {
		var ruleset Pfioc_ruleset

		ruleset.Nr = i

		err := p.ioctl.Read(DIOCGETRULESET, unsafe.Pointer(&ruleset))
		if err != nil {
			return nil, fmt.Errorf("get ruleset: %s", err)
		}

		if ruleset.NameString() == name {
			return &ruleset, nil
		}
	}

	return nil, nil
}

type LookUpNATOpts struct {
	SrcIP   net.IP
	SrcPort int32
	DstIP   net.IP
	DstPort int32
}

func (p *Pfctl) LookUpNAT(opts LookUpNATOpts) (net.IP, int, error) {
	var natlook Pfioc_natlook

	natlook.SetSrcIP(opts.SrcIP)
	natlook.SetSrcPort(opts.SrcPort)
	natlook.SetDstIP(opts.DstIP)
	natlook.SetDstPort(opts.DstPort)
	natlook.af = syscall.AF_INET
	natlook.proto = syscall.IPPROTO_TCP
	natlook.direction = PF_OUT

	err := p.ioctl.Read(DIOCNATLOOK, unsafe.Pointer(&natlook))
	if err != nil {
		return nil, 0, fmt.Errorf("natlook: %s", err)
	}

	return natlook.GetIP(), natlook.GetPort(), nil
}
