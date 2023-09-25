package net

import (
	"net"

	"github.com/carvel-dev/kwt/pkg/kwt/net/dstconn"
)

type EntryPoint interface {
	EntryPoint() (EntryPointSession, error)
	Delete() error
}

type EntryPointSession interface {
	Opts() dstconn.SSHClientConnOpts
	Close() error
}

type Subnets interface {
	Subnets() ([]net.IPNet, error)
}

type DNSIPs interface {
	DNSIPs() ([]net.IP, error)
}

type DNSServerFactory interface {
	NewDNSServer(dstconn.Factory) (DNSServer, error)
	NewDNSOSCache() DNSOSCache
}

type DNSServer interface {
	Serve(startedCh chan struct{}) error
	TCPAddr() net.Addr
	UDPAddr() net.Addr
	Shutdown() error
}
