package dstconn

import (
	"net"
)

type ClosingConnCopier struct{}

var _ ConnCopier = ClosingConnCopier{}

func (c ClosingConnCopier) CopyAndClose(dstConn net.Conn, srcConn net.Conn) {
	dstConn.Close()
	srcConn.Close()
}
