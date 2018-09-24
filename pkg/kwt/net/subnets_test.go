package net_test

import (
	"net"
	"testing"

	. "github.com/cppforlife/kwt/pkg/kwt/net"
)

func TestGuessSubnets(t *testing.T) {
	excludedIPs := []net.IP{net.ParseIP("10.80.130.76")}

	inputIPs := []net.IP{net.ParseIP("10.200.37.33"), net.ParseIP("10.100.200.141")}
	subnets := GuessSubnets(inputIPs, excludedIPs)
	if SubnetsAsString(subnets) != "10.200.37.33/14, 10.100.200.141/14" {
		t.Fatalf("did not guess subnets correctly: %s", SubnetsAsString(subnets))
	}

	inputIPs = []net.IP{net.ParseIP("10.200.37.33"), net.ParseIP("10.200.100.141")}
	subnets = GuessSubnets(inputIPs, excludedIPs)
	if SubnetsAsString(subnets) != "10.200.37.33/14" {
		t.Fatalf("did not guess subnets correctly: %s", SubnetsAsString(subnets))
	}
}
