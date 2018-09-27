package forwarder

import (
	"fmt"
	"net"
	"strconv"
)

type Iptables struct {
	opts   IptablesOpts
	chains []chain

	exec CmdExecutor

	logTag string
	logger Logger
}

var _ Forwarder = Iptables{}

type IptablesOpts struct {
	DstTCPPort     int
	DstDNSTCPPort  int
	DstDNSUDPPort  int
	ProcessGroupID int
}

type chain struct {
	Name       string
	Type       string
	GroupCheck []string
}

func NewIptables(opts IptablesOpts, exec CmdExecutor, logger Logger) Iptables {
	name := fmt.Sprintf("kwt-tcp-%d", opts.DstTCPPort)
	chains := []chain{
		{
			Name:       name + "-output",
			Type:       "OUTPUT",
			GroupCheck: []string{"-m", "owner", "!", "--gid-owner", strconv.Itoa(opts.ProcessGroupID)}},
		{
			Name: name + "-prerouting",
			Type: "PREROUTING",
		},
	}

	return Iptables{opts, chains, exec, "Iptables", logger}
}

func (i Iptables) CheckPrereqs() error {
	out, err := i.runCmd([]string{"-L", "-t", "nat"})
	if err != nil {
		return fmt.Errorf("Checking 'iptables' can run successfully: %s (output: %s)", err, out)
	}

	return nil
}

func (i Iptables) Add(subnets []net.IPNet, dnsIPs []net.IP) error {
	cmds := [][]string{}

	for _, chain := range i.chains {
		cmds = append(cmds, [][]string{
			[]string{"-t", "nat", "-N", chain.Name},
			[]string{"-t", "nat", "-F", chain.Name},
			[]string{"-t", "nat", "-I", chain.Type, "1", "-j", chain.Name},
		}...)

		for _, subnet := range subnets {
			cmds = append(cmds, append([]string{
				"-t", "nat", "-A", chain.Name,
				"-j", "REDIRECT", "--dest", subnet.String(), "-p", "tcp",
				"--to-ports", strconv.Itoa(i.opts.DstTCPPort),
				"-m", "ttl", "!", "--ttl", "42",
			}, chain.GroupCheck...))
		}

		for _, ip := range dnsIPs {
			cmds = append(cmds, append([]string{
				"-t", "nat", "-A", chain.Name,
				"-j", "REDIRECT", "--dest", ip.String() + "/32", "-p", "tcp",
				"--dport", "53", "--to-ports", strconv.Itoa(i.opts.DstDNSTCPPort),
				"-m", "ttl", "!", "--ttl", "42",
			}, chain.GroupCheck...))

			cmds = append(cmds, append([]string{
				"-t", "nat", "-A", chain.Name,
				"-j", "REDIRECT", "--dest", ip.String() + "/32", "-p", "udp",
				"--dport", "53", "--to-ports", strconv.Itoa(i.opts.DstDNSUDPPort),
				"-m", "ttl", "!", "--ttl", "42",
			}, chain.GroupCheck...))
		}

		cmds = append(cmds, []string{
			"-t", "nat", "-A", chain.Name,
			"-j", "RETURN", "--dest", "127.0.0.0/8", "-p", "tcp",
		})
	}

	return i.runCmds(cmds)
}

func (i Iptables) Reset() error {
	cmds := [][]string{}

	for _, chain := range i.chains {
		cmds = append(cmds, [][]string{
			[]string{"-t", "nat", "-D", chain.Type, "-j", chain.Name},
			[]string{"-t", "nat", "-F", chain.Name},
			[]string{"-t", "nat", "-X", chain.Name},
		}...)
	}

	return i.runCmds(cmds)
}

func (i Iptables) runCmds(cmds [][]string) error {
	for _, cmd := range cmds {
		_, err := i.runCmd(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i Iptables) runCmd(cmd []string) ([]byte, error) {
	return i.exec.CombinedOutput("iptables", append([]string{"-w"}, cmd...), nil)
}
