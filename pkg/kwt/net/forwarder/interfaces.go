package forwarder

import (
	"net"
)

type Forwarder interface {
	CheckPrereqs() error
	Add([]net.IPNet, []net.IP) error
	Reset() error
}

type OriginalDstResolver interface {
	GetOrigIPPort(net.Conn) (net.IP, int, error)
}
