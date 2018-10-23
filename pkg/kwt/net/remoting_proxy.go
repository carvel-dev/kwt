package net

import (
	"fmt"
)

type RemotingProxy struct {
	entryPoint      EntryPoint
	subnets         Subnets
	dnsIPs          DNSIPs
	forwardingProxy *ForwardingProxy

	logTag string
	logger Logger
}

func NewRemotingProxy(entryPoint EntryPoint, subnets Subnets, dnsIPs DNSIPs, forwardingProxy *ForwardingProxy, logger Logger) *RemotingProxy {
	return &RemotingProxy{entryPoint, subnets, dnsIPs, forwardingProxy, "RemotingProxy", logger}
}

func (f *RemotingProxy) Serve() error {
	subnets, err := f.subnets.Subnets()
	if err != nil {
		return err
	}

	if len(subnets) == 0 {
		return fmt.Errorf("Expected at least one subnet to be guessed or specified")
	}

	dnsIPs, err := f.dnsIPs.DNSIPs()
	if err != nil {
		return err
	}

	reconnSSHClient := NewReconnSSHClient(f.entryPoint, f.logger)

	// Connect immediately starting proxies to catch any early failures
	err = reconnSSHClient.Connect()
	if err != nil {
		return err
	}

	defer reconnSSHClient.Disconnect()

	return f.forwardingProxy.Serve(reconnSSHClient, subnets, dnsIPs)
}

func (f *RemotingProxy) Shutdown() error {
	return f.forwardingProxy.Shutdown()
}
