package pf

import (
	"net"
	"unsafe"
)

const Pfioc_natlookSize = 84

type Pfioc_natlook struct {
	saddr  [4]uint32 // network byte order
	daddr  [4]uint32 // network byte order
	rsaddr [4]uint32 // network byte order
	rdaddr [4]uint32 // network byte order

	sxport  [2]uint16 // network byte order
	dxport  [2]uint16 // network byte order
	rsxport [2]uint16 // network byte order
	rdxport [2]uint16 // network byte order

	af            uint8
	proto         uint8
	proto_variant uint8
	direction     uint8
}

func (natlook *Pfioc_natlook) SetSrcIP(ip net.IP) {
	natlook.saddr[0] = Htonl(natlook.ipToL(ip))
}

func (natlook *Pfioc_natlook) SetSrcPort(port int32) {
	natlook.sxport[0] = Htons(uint16(port))
}

func (natlook *Pfioc_natlook) SetDstIP(ip net.IP) {
	natlook.daddr[0] = Htonl(natlook.ipToL(ip))
}

func (natlook *Pfioc_natlook) SetDstPort(port int32) {
	natlook.dxport[0] = Htons(uint16(port))
}

func (natlook *Pfioc_natlook) GetIP() net.IP {
	return natlook.lToIP(Ntohl(natlook.rdaddr[0]))
}

func (natlook *Pfioc_natlook) GetPort() int {
	return int(Ntohs(natlook.rdxport[0]))
}

func (*Pfioc_natlook) ipToL(ip net.IP) uint32 {
	ip = ip.To4() // turns into 4 byte slice
	var a, b, c, d uint32 = uint32(ip[0]), uint32(ip[1]), uint32(ip[2]), uint32(ip[3])
	return a<<24 | b<<16 | c<<8 | d
}

func (*Pfioc_natlook) lToIP(l uint32) net.IP {
	return net.IP([]byte{
		byte(l >> 24 & 255),
		byte(l >> 16 & 255),
		byte(l >> 8 & 255),
		byte(l & 255),
	})
}

func init() {
	var natlook Pfioc_natlook
	if unsafe.Sizeof(natlook) != Pfioc_natlookSize {
		panic("Expected Pfioc_natlook to match")
	}
}
