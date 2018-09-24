package net

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

func LocalIPs() ([]net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("Fetching local interface addrs: %s", err)
	}

	var localIPv4s []net.IP

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			return nil, err
		}

		ipv4 := ip.To4()
		if ipv4 != nil {
			if !ipv4.IsLoopback() {
				localIPv4s = append(localIPv4s, ipv4)
			}
		}
	}

	return localIPv4s, nil
}

func GuessSubnets(remoteIPs []net.IP, excludedIPs []net.IP) []net.IPNet {
	largestToSmallestMasks := []net.IPMask{
		net.CIDRMask(14, 32), // default on GKE
		net.CIDRMask(16, 32),
		net.CIDRMask(24, 32),
	}
	var possibleNets []net.IPNet

	// Guess ranges that would satsify remote IPs but not overlap with local IPs
	for _, remoteIP := range remoteIPs {
		for _, mask := range largestToSmallestMasks {
			possibleNet := net.IPNet{remoteIP, mask}

			overlapsLocal := false
			for _, excludedIP := range excludedIPs {
				if possibleNet.Contains(excludedIP) {
					overlapsLocal = true
				}
			}

			if !overlapsLocal {
				possibleNets = append(possibleNets, possibleNet)
				break // continue with the next remote IP
			}
		}
	}

	// Sort largest subnet first
	sort.Slice(possibleNets, func(i, j int) bool {
		iOnes, _ := possibleNets[i].Mask.Size()
		jOnes, _ := possibleNets[j].Mask.Size()
		return iOnes < jOnes
	})

	var selectedNets []net.IPNet
	possibleNetsIdxAdded := map[int]struct{}{}

	for _, remoteIP := range remoteIPs {
		for ni, n := range possibleNets {
			if n.Contains(remoteIP) {
				if _, found := possibleNetsIdxAdded[ni]; !found {
					possibleNetsIdxAdded[ni] = struct{}{}
					selectedNets = append(selectedNets, n)
				}
				break
			}
		}
	}

	return selectedNets
}

func SubnetsAsString(subnets []net.IPNet) string {
	var result []string
	for _, n := range subnets {
		result = append(result, n.String())
	}
	return strings.Join(result, ", ")
}
