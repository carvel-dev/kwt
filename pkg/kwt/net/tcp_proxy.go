package net

import (
	"fmt"
	"net"
	"time"

	"github.com/cppforlife/kwt/pkg/kwt/net/dstconn"
	"github.com/cppforlife/kwt/pkg/kwt/net/forwarder"
)

type TCPProxy struct {
	origDstResolver forwarder.OriginalDstResolver
	dstConnFactory  dstconn.Factory

	listener net.Listener

	logTag string
	logger Logger
}

func NewTCPProxy(
	origDstResolver forwarder.OriginalDstResolver,
	dstConnFactory dstconn.Factory,
	logger Logger,
) *TCPProxy {
	return &TCPProxy{
		origDstResolver: origDstResolver,
		dstConnFactory:  dstConnFactory,

		logTag: "TCPProxy",
		logger: logger,
	}
}

func (c *TCPProxy) Serve(startedCh chan struct{}) error {
	var err error

	c.listener, err = net.Listen("tcp", "localhost:0")
	if err != nil {
		return err
	}

	defer c.listener.Close()

	startedCh <- struct{}{}

	c.logger.Info(c.logTag, "Started proxy on %s", c.Addr())

	for {
		conn, err := c.listener.Accept()
		if err != nil {
			return err
		}
		go c.serveConn(conn)
	}
}

func (c *TCPProxy) Addr() net.Addr { return c.listener.Addr() }

func (c *TCPProxy) Shutdown() error {
	// TODO drain connections?
	return nil
}

func (c *TCPProxy) serveConn(srcConn net.Conn) {
	t1 := time.Now()

	srcDesc := srcConn.RemoteAddr()
	c.logger.Info(c.logTag, "Received %s", srcDesc)

	origDstIP, origDstPort, err := c.origDstResolver.GetOrigIPPort(srcConn)
	if err != nil {
		c.logger.Error(c.logTag, "Could not retrieve original destination from '%s': %s", srcDesc, err)
		srcConn.Close()
		return
	}

	dstDesc := fmt.Sprintf("%s:%d", origDstIP, origDstPort)

	dstConn, err := c.dstConnFactory.NewConn(origDstIP, origDstPort)
	if err != nil {
		c.logger.Error(c.logTag, "Could not establish remote connection to '%s': %s", dstDesc, err)
		srcConn.Close()
		return
	}

	t2 := time.Now()

	proxyDesc := fmt.Sprintf("%s->%s", srcDesc, dstDesc)
	c.logger.Info(c.logTag, "Started %s", proxyDesc)

	defer func() {
		t3 := time.Now()
		c.logger.Info(c.logTag, "Finished %s (%s/%s)", proxyDesc, t2.Sub(t1), t3.Sub(t2))
	}()

	c.dstConnFactory.NewConnCopier(proxyDesc).CopyAndClose(dstConn, srcConn)
}
