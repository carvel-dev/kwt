package net

import (
	"fmt"

	"github.com/cppforlife/kwt/pkg/kwt/net/dstconn"
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
	sshConnOpts, err := f.entryPoint.EntryPoint()
	if err != nil {
		return err
	}

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

	sshClient := dstconn.NewSSHClient(sshConnOpts, f.logger)

	err = sshClient.Connect()
	if err != nil {
		return err
	}

	defer func() {
		err = sshClient.Disconnect()
		if err != nil {
			f.logger.Error(f.logTag, "Failed disconnecting SSH client: %s", err)
		}
	}()

	return f.forwardingProxy.Serve(sshClient, subnets, dnsIPs)
}

func (f *RemotingProxy) Shutdown() error {
	return f.forwardingProxy.Shutdown()
}
