package dns

import (
	"fmt"
	"net"
)

type MultiIPResolver struct {
	resolvers []IPResolver
}

var _ IPResolver = MultiIPResolver{}

func NewMultiIPResolver(resolvers []IPResolver) MultiIPResolver {
	return MultiIPResolver{resolvers}
}

func (r MultiIPResolver) ResolveIPv4(question string) ([]net.IP, bool, error) {
	for _, resolver := range r.resolvers {
		ips, resolved, err := resolver.ResolveIPv4(question)
		if resolved || err != nil {
			return ips, resolved, err
		}
	}

	return nil, false, fmt.Errorf("Could not resolve IPv4")
}

func (r MultiIPResolver) ResolveIPv6(question string) ([]net.IP, bool, error) {
	for _, resolver := range r.resolvers {
		ips, resolved, err := resolver.ResolveIPv6(question)
		if resolved || err != nil {
			return ips, resolved, err
		}
	}

	return nil, false, fmt.Errorf("Could not resolve IPv6")
}
