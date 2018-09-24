package main

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/cppforlife/kwt/pkg/kwt/net/pf"
)

func main() {
	err := run()
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	ioctl, err := pf.NewIoctl("/dev/pf")
	if err != nil {
		return fmt.Errorf("opening /dev/pf: %s", err)
	}

	defer ioctl.Close()

	return list(ioctl)
}

func list(ioctl *pf.Ioctl) error {
	var ruleset pf.Pfioc_ruleset

	err := ioctl.Read(pf.DIOCGETRULESETS, unsafe.Pointer(&ruleset))
	if err != nil {
		return fmt.Errorf("get rulesets: %s", err)
	}

	fmt.Printf("# of anchors: %d\n", ruleset.Nr)

	maxLen := ruleset.Nr

	var i uint32

	for ; i < maxLen; i++ {
		var ruleset pf.Pfioc_ruleset

		ruleset.Nr = i

		err := ioctl.Read(pf.DIOCGETRULESET, unsafe.Pointer(&ruleset))
		if err != nil {
			return fmt.Errorf("get ruleset: %s", err)
		}

		if ruleset.NameString() == os.Args[1] {
			err := del(ioctl, os.Args[1], ruleset.Nr, pf.PF_PASS)
			if err != nil {
				return fmt.Errorf("deleting pf-pass: %s", err)
			}

			err = del(ioctl, os.Args[1], ruleset.Nr, pf.PF_RDR)
			if err != nil {
				return fmt.Errorf("deleting pf-rdr: %s", err)
			}

			return nil
		}

		fmt.Printf("anchor: nr='%d' path='%s' name='%s'\n", ruleset.Nr, ruleset.Path, ruleset.Name)
	}

	return nil
}

func del(ioctl *pf.Ioctl, name string, nr uint32, ruleAction pf.RuleAction) error {
	var rule pf.Pfioc_rule

	fmt.Printf("deleting...\n")

	rule.SetAction(pf.PF_CHANGE_GET_TICKET)
	rule.Nr = nr
	rule.SetRuleAction(ruleAction)

	err := ioctl.Read(pf.DIOCCHANGERULE, unsafe.Pointer(&rule))
	if err != nil {
		return fmt.Errorf("change rule (get ticket): %s", err)
	}

	rule.SetAction(pf.PF_CHANGE_REMOVE)

	err = ioctl.Read(pf.DIOCCHANGERULE, unsafe.Pointer(&rule))
	if err != nil {
		return fmt.Errorf("change rule (remove): %s", err)
	}

	return nil
}
