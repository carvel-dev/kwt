package forwarder

import (
	"net"
)

type StaticResolver struct {
	ip   net.IP
	port int
}

var _ OriginalDstResolver = StaticResolver{}

func NewStaticResolver(ip net.IP, port int) StaticResolver {
	return StaticResolver{ip, port}
}

func (r StaticResolver) GetOrigIPPort(conn net.Conn) (net.IP, int, error) {
	return r.ip, r.port, nil
}
