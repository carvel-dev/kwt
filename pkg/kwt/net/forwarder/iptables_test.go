package forwarder_test

import (
	"io"
	"net"
	"reflect"
	"testing"

	. "github.com/cppforlife/kwt/pkg/kwt/net/forwarder"
)

type FakeCmdExecutor struct {
	Cmds [][]string
}

var _ CmdExecutor = &FakeCmdExecutor{}

func (e *FakeCmdExecutor) CombinedOutput(cmdName string, args []string, stdin io.Reader) ([]byte, error) {
	e.Cmds = append(e.Cmds, append([]string{cmdName}, args...))
	return nil, nil
}

func TestIptables(t *testing.T) {
	exec := &FakeCmdExecutor{}
	opts := IptablesOpts{
		DstTCPPort:     123,
		DstDNSTCPPort:  200,
		DstDNSUDPPort:  124,
		ProcessGroupID: 125,
	}
	iptables := NewIptables(opts, exec, nil)

	_, ipNet1, _ := net.ParseCIDR("10.0.0.0/24")
	_, ipNet2, _ := net.ParseCIDR("192.0.0.0/24")

	ip1 := net.ParseIP("1.1.1.1")
	ip2 := net.ParseIP("2.2.2.2")

	err := iptables.Add([]net.IPNet{*ipNet1, *ipNet2}, []net.IP{ip1, ip2})
	if err != nil {
		t.Fatalf("Expected no err: %s", err)
	}

	expectedCmds := [][]string{
		[]string{"iptables", "-w", "-t", "nat", "-N", "kwt-tcp-123-output"},
		[]string{"iptables", "-w", "-t", "nat", "-F", "kwt-tcp-123-output"},
		[]string{"iptables", "-w", "-t", "nat", "-I", "OUTPUT", "1", "-j", "kwt-tcp-123-output"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-output", "-j", "REDIRECT", "--dest", "10.0.0.0/24", "-p", "tcp", "--to-ports", "123", "-m", "ttl", "!", "--ttl", "42", "-m", "owner", "!", "--gid-owner", "125"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-output", "-j", "REDIRECT", "--dest", "192.0.0.0/24", "-p", "tcp", "--to-ports", "123", "-m", "ttl", "!", "--ttl", "42", "-m", "owner", "!", "--gid-owner", "125"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-output", "-j", "REDIRECT", "--dest", "1.1.1.1/32", "-p", "tcp", "--dport", "53", "--to-ports", "200", "-m", "ttl", "!", "--ttl", "42", "-m", "owner", "!", "--gid-owner", "125"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-output", "-j", "REDIRECT", "--dest", "1.1.1.1/32", "-p", "udp", "--dport", "53", "--to-ports", "124", "-m", "ttl", "!", "--ttl", "42", "-m", "owner", "!", "--gid-owner", "125"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-output", "-j", "REDIRECT", "--dest", "2.2.2.2/32", "-p", "tcp", "--dport", "53", "--to-ports", "200", "-m", "ttl", "!", "--ttl", "42", "-m", "owner", "!", "--gid-owner", "125"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-output", "-j", "REDIRECT", "--dest", "2.2.2.2/32", "-p", "udp", "--dport", "53", "--to-ports", "124", "-m", "ttl", "!", "--ttl", "42", "-m", "owner", "!", "--gid-owner", "125"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-output", "-j", "RETURN", "--dest", "127.0.0.0/8", "-p", "tcp"},

		[]string{"iptables", "-w", "-t", "nat", "-N", "kwt-tcp-123-prerouting"},
		[]string{"iptables", "-w", "-t", "nat", "-F", "kwt-tcp-123-prerouting"},
		[]string{"iptables", "-w", "-t", "nat", "-I", "PREROUTING", "1", "-j", "kwt-tcp-123-prerouting"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-prerouting", "-j", "REDIRECT", "--dest", "10.0.0.0/24", "-p", "tcp", "--to-ports", "123", "-m", "ttl", "!", "--ttl", "42"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-prerouting", "-j", "REDIRECT", "--dest", "192.0.0.0/24", "-p", "tcp", "--to-ports", "123", "-m", "ttl", "!", "--ttl", "42"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-prerouting", "-j", "REDIRECT", "--dest", "1.1.1.1/32", "-p", "tcp", "--dport", "53", "--to-ports", "200", "-m", "ttl", "!", "--ttl", "42"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-prerouting", "-j", "REDIRECT", "--dest", "1.1.1.1/32", "-p", "udp", "--dport", "53", "--to-ports", "124", "-m", "ttl", "!", "--ttl", "42"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-prerouting", "-j", "REDIRECT", "--dest", "2.2.2.2/32", "-p", "tcp", "--dport", "53", "--to-ports", "200", "-m", "ttl", "!", "--ttl", "42"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-prerouting", "-j", "REDIRECT", "--dest", "2.2.2.2/32", "-p", "udp", "--dport", "53", "--to-ports", "124", "-m", "ttl", "!", "--ttl", "42"},
		[]string{"iptables", "-w", "-t", "nat", "-A", "kwt-tcp-123-prerouting", "-j", "RETURN", "--dest", "127.0.0.0/8", "-p", "tcp"},
	}

	if !reflect.DeepEqual(exec.Cmds, expectedCmds) {
		t.Fatalf("Expected Add cmds to match: actual %#v", exec.Cmds)
	}

	exec.Cmds = nil

	err = iptables.Reset()
	if err != nil {
		t.Fatalf("Expected no err: %s", err)
	}

	expectedCmds = [][]string{
		[]string{"iptables", "-w", "-t", "nat", "-D", "OUTPUT", "-j", "kwt-tcp-123-output"},
		[]string{"iptables", "-w", "-t", "nat", "-F", "kwt-tcp-123-output"},
		[]string{"iptables", "-w", "-t", "nat", "-X", "kwt-tcp-123-output"},

		[]string{"iptables", "-w", "-t", "nat", "-D", "PREROUTING", "-j", "kwt-tcp-123-prerouting"},
		[]string{"iptables", "-w", "-t", "nat", "-F", "kwt-tcp-123-prerouting"},
		[]string{"iptables", "-w", "-t", "nat", "-X", "kwt-tcp-123-prerouting"},
	}

	if !reflect.DeepEqual(exec.Cmds, expectedCmds) {
		t.Fatalf("Expected Reset cmds to match: actual %#v", exec.Cmds)
	}
}
