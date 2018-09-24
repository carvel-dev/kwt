package dns

import (
	"net"
)

type StaticIPResolver struct {
	ip net.IP
}

var _ IPResolver = StaticIPResolver{}

func NewStaticIPResolver(ip net.IP) StaticIPResolver {
	return StaticIPResolver{ip}
}

func (r StaticIPResolver) ResolveIPv4(question string) ([]net.IP, bool, error) {
	return []net.IP{r.ip}, true, nil
}

func (r StaticIPResolver) ResolveIPv6(question string) ([]net.IP, bool, error) {
	return nil, true, nil
}
