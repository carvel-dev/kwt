package dns

import (
	"net"
)

type StaticIPsResolver struct {
	ips []net.IP
}

var _ IPResolver = StaticIPsResolver{}

func NewStaticIPsResolver(ips []net.IP) StaticIPsResolver {
	return StaticIPsResolver{ips}
}

func (r StaticIPsResolver) ResolveIPv4(question string) ([]net.IP, bool, error) {
	return r.ips, true, nil
}

func (r StaticIPsResolver) ResolveIPv6(question string) ([]net.IP, bool, error) {
	return nil, true, nil
}
