package forwarder

import (
	"net"
)

type Forwarder interface {
	Add([]net.IPNet, []net.IP) error
	Reset() error
}

type OriginalDstResolver interface {
	GetOrigIPPort(net.Conn) (net.IP, int, error)
}
