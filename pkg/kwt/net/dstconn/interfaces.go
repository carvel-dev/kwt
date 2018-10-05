package dstconn

import (
	"net"
)

type Factory interface {
	NewConn(net.IP, int) (net.Conn, error)
	NewConnCopier(logTag string) ConnCopier

	NewListener() (net.Listener, error)
}

type ConnCopier interface {
	CopyAndClose(dstConn net.Conn, srcConn net.Conn)
}
