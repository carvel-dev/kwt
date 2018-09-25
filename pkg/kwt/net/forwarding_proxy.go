package net

import (
	"fmt"
	"net"
	"strconv"

	"github.com/cppforlife/kwt/pkg/kwt/net/dstconn"
	"github.com/cppforlife/kwt/pkg/kwt/net/forwarder"
)

type ForwardingProxy struct {
	forwarderFactory forwarder.Factory
	dnsServerFactory DNSServerFactory
	shutdownCh       chan struct{}

	logTag string
	logger Logger
}

func NewForwardingProxy(forwarderFactory forwarder.Factory, dnsServerFactory DNSServerFactory, logger Logger) *ForwardingProxy {
	return &ForwardingProxy{forwarderFactory, dnsServerFactory, make(chan struct{}), "ForwardingProxy", logger}
}

func (o *ForwardingProxy) Serve(dstConnFactory dstconn.Factory, subnets []net.IPNet, dnsIPs []net.IP) error {
	origDstResolver, err := o.forwarderFactory.NewOriginalDstResolver()
	if err != nil {
		return err
	}

	tcpProxy := NewTCPProxy(origDstResolver, dstConnFactory, o.logger)
	tcpProxyErrCh := make(chan error)
	tcpProxyStartedCh := make(chan struct{})

	udpProxy := NewUDPProxy(dstConnFactory, o.logger)
	udpProxyErrCh := make(chan error)
	udpProxyStartedCh := make(chan struct{})

	dnsServer, err := o.dnsServerFactory.NewDNSServer(dstConnFactory)
	if err != nil {
		return err
	}

	dnsServerErrCh := make(chan error)
	dnsServerStartedCh := make(chan struct{})

	forwarder := forwarder.NewLocking()
	forwarderErrCh := make(chan error)

	go func() {
		tcpProxyErrCh <- tcpProxy.Serve(tcpProxyStartedCh)
	}()

	go func() {
		udpProxyErrCh <- udpProxy.Serve(udpProxyStartedCh)
	}()

	go func() {
		dnsServerErrCh <- dnsServer.Serve(dnsServerStartedCh)
	}()

	go func() {
		<-tcpProxyStartedCh
		<-udpProxyStartedCh
		<-dnsServerStartedCh

		actualForwarder, err := o.buildForwarder(tcpProxy.Addr(), dnsServer.UDPAddr())
		if err != nil {
			forwarderErrCh <- err
			return
		}

		err = actualForwarder.CheckPrereqs()
		if err != nil {
			forwarderErrCh <- err
			return
		}

		forwarder.SetForwarder(actualForwarder)

		o.logger.Info(o.logTag, "Forwarding subnets: %s", SubnetsAsString(subnets))

		err = forwarder.Add(subnets, dnsIPs)
		if err != nil {
			forwarderErrCh <- err
			return
		}

		o.logger.Info(o.logTag, "Ready!")
	}()

	errCh := make(chan error)

	go func() {
		select {
		case <-o.shutdownCh:
			errCh <- nil
		case err := <-tcpProxyErrCh:
			errCh <- err
		case err := <-udpProxyErrCh:
			errCh <- err
		case err := <-dnsServerErrCh:
			errCh <- err
		case err := <-forwarderErrCh:
			errCh <- err
		}
	}()

	origErr := <-errCh

	// Most import to reset since forwarder's resources are not reclaimed automatically
	err = forwarder.Reset()
	if err != nil {
		o.logger.Error(o.logTag, "Failed resetting forwarder: %s", err)
	}

	err = tcpProxy.Shutdown()
	if err != nil {
		o.logger.Error(o.logTag, "Failed shutting down TCP proxy: %s", err)
	}

	err = udpProxy.Shutdown()
	if err != nil {
		o.logger.Error(o.logTag, "Failed shutting down UDP proxy: %s", err)
	}

	err = dnsServer.Shutdown()
	if err != nil {
		o.logger.Error(o.logTag, "Failed shutting down DNS server: %s", err)
	}

	return origErr
}

func (o *ForwardingProxy) Shutdown() error {
	close(o.shutdownCh)
	return nil
}

func (o *ForwardingProxy) buildForwarder(tcpAddr, dnsUDPAddr net.Addr) (forwarder.Forwarder, error) {
	tcpPort, err := o.portFromAddr(tcpAddr)
	if err != nil {
		return nil, err
	}

	dnsUDPPort, err := o.portFromAddr(dnsUDPAddr)
	if err != nil {
		return nil, err
	}

	return o.forwarderFactory.NewForwarder(forwarder.ForwarderOpts{
		DstTCPPort:    tcpPort,
		DstDNSUDPPort: dnsUDPPort,
	})
}

func (*ForwardingProxy) portFromAddr(addr net.Addr) (int, error) {
	_, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		return 0, fmt.Errorf("Parsing addr: %s", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("Parsing port: %s", err)
	}

	return port, nil
}
