package net

import (
	"net"

	"github.com/cppforlife/kwt/pkg/kwt/net/dstconn"
	"github.com/miekg/dns"
)

// https://github.com/LiamHaworth/go-tproxy/blob/master/tproxy_udp.go

type UDPProxy struct {
	dstConnFactory dstconn.Factory

	conn *net.UDPConn

	logTag string
	logger Logger
}

func NewUDPProxy(
	dstConnFactory dstconn.Factory,
	logger Logger,
) *UDPProxy {
	return &UDPProxy{
		dstConnFactory: dstConnFactory,

		logTag: "UDPProxy",
		logger: logger,
	}
}

func (c *UDPProxy) Serve(startedCh chan struct{}) error {
	var err error

	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		return err
	}

	c.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	defer c.conn.Close()

	startedCh <- struct{}{}

	c.logger.Info(c.logTag, "Started proxy on %s", c.conn.LocalAddr())

	for {
		var b []byte
		_, sess, err := dns.ReadFromSessionUDP(c.conn, b)
		if err != nil {
			c.logger.Error(c.logTag, "Receiving UDP", err)
		} else {
			go c.serveSess(sess)
		}
	}
}

func (c *UDPProxy) Addr() net.Addr { return c.conn.LocalAddr() }

func (c *UDPProxy) Shutdown() error {
	return nil
}

func (c *UDPProxy) serveSess(sess *dns.SessionUDP) {
	c.logger.Info(c.logTag, "Started %s", sess.RemoteAddr())
}
