package forwarder

// Taken from https://play.golang.org/p/GMAaKucHOr
// Realated: https://github.com/schmurfy/squid/blob/a33e4089e463d175014d57afed74d3897a7de02d/src/ip/IpIntercept.cc

import (
	"errors"
	"net"
	"syscall"
	"unsafe"
)

type LinuxOriginalDstResolver struct{}

var _ OriginalDstResolver = LinuxOriginalDstResolver{}

type sockaddr struct {
	family uint16
	data   [14]byte
}

func (r LinuxOriginalDstResolver) GetOrigIPPort(conn net.Conn) (net.IP, int, error) {
	tcpConn, ok := (conn).(*net.TCPConn)
	if !ok {
		return nil, 0, errors.New("not a TCPConn")
	}

	file, err := tcpConn.File()
	if err != nil {
		return nil, 0, err
	}

	// TODO to avoid potential problems from making the socket non-blocking. huh?
	// tcpConn.Close()

	// todo do we need this?
	// *conn, err = net.FileConn(file)
	// if err != nil {
	//  return "", 0, err
	// }

	defer file.Close()
	fd := file.Fd()

	const soOriginalDst = 80

	var addr sockaddr
	size := uint32(unsafe.Sizeof(addr))

	err = r.getsockopt(int(fd), solIP, soOriginalDst, uintptr(unsafe.Pointer(&addr)), &size)
	if err != nil {
		return nil, 0, err
	}

	var ip net.IP
	switch addr.family {
	case syscall.AF_INET:
		ip = addr.data[2:6]
	default:
		return nil, 0, errors.New("unrecognized address family")
	}

	port := int(addr.data[0])<<8 + int(addr.data[1])

	return ip, port, nil
}

func (LinuxOriginalDstResolver) getsockopt(s int, level int, name int, val uintptr, vallen *uint32) (err error) {
	_, _, e1 := syscall.Syscall6(syscall.SYS_GETSOCKOPT, uintptr(s), uintptr(level), uintptr(name), uintptr(val), uintptr(unsafe.Pointer(vallen)), 0)
	if e1 != 0 {
		err = e1
	}
	return
}
