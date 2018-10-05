package dstconn

import (
	"net"
)

type Local struct {
	logger Logger
}

var _ Factory = Local{}

func NewLocal(logger Logger) Local {
	return Local{logger}
}

func (l Local) NewConn(ip net.IP, port int) (net.Conn, error) {
	return net.DialTCP("tcp", nil, &net.TCPAddr{ip, port, ""})
}

func (l Local) NewConnCopier(logTag string) ConnCopier {
	return NewSSHConnCopier(logTag, l.logger) // TODO plain one?
}

func (l Local) NewListener() (net.Listener, error) {
	panic("Not implemented")
}
