package mdns

import (
	"net"
)

const Domain = "local."

var (
	mdnsIPv4Addr = &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 251),
		Port: 5353,
	}
)
