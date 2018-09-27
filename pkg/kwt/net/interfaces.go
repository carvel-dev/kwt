package net

import (
	"net"

	"github.com/cppforlife/kwt/pkg/kwt/net/dstconn"
)

type EntryPoint interface {
	EntryPoint() (dstconn.SSHClientConnOpts, error)
	Delete() error
}

type Subnets interface {
	Subnets() ([]net.IPNet, error)
}

type DNSIPs interface {
	DNSIPs() ([]net.IP, error)
}

type DNSServerFactory interface {
	NewDNSServer(dstconn.Factory) (DNSServer, error)
}

type DNSServer interface {
	Serve(startedCh chan struct{}) error
	TCPAddr() net.Addr
	UDPAddr() net.Addr
	Shutdown() error
}
