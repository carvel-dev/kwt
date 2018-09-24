package net

import (
	"net"
)

type ConfiguredSubnets struct {
	subnets []string
}

func NewConfiguredSubnets(subnets []string) ConfiguredSubnets {
	return ConfiguredSubnets{subnets}
}

func (s ConfiguredSubnets) Subnets() ([]net.IPNet, error) {
	var result []net.IPNet

	for _, subnet := range s.subnets {
		_, ipNet, err := net.ParseCIDR(subnet)
		if err != nil {
			return nil, err
		}
		result = append(result, *ipNet)
	}

	return result, nil
}
