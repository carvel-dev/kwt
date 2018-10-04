package net

import (
	"fmt"
	"io"
	"net"
	"sync"

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

type ReconnSSHClient struct {
	entryPoint     EntryPoint
	entryPointSess EntryPointSession

	sshClient     *dstconn.SSHClient
	sshClientLock sync.RWMutex

	logger Logger
	logTag string
}

func NewReconnSSHClient(entryPoint EntryPoint, logger Logger) *ReconnSSHClient {
	return &ReconnSSHClient{
		entryPoint: entryPoint,
		logger:     logger,
		logTag:     "ReconnSSHClient",
	}
}

var _ dstconn.Factory = &ReconnSSHClient{}

func (f *ReconnSSHClient) NewConn(ip net.IP, port int) (net.Conn, error) {
	client, err := f.getSSHClient()
	if err != nil {
		return nil, err
	}

	conn, err := client.NewConn(ip, port)
	if err != nil {
		isEOF := err == io.EOF
		f.logger.Debug(f.logTag, "Received err: %s (isEOF: %t)", err, isEOF)

		if isEOF {
			f.disconnect()

			client, err := f.connect()
			if err != nil {
				return nil, err
			}

			return client.NewConn(ip, port)
		}

		return nil, err
	}

	return conn, nil
}

func (f *ReconnSSHClient) NewConnCopier(proxyDesc string) dstconn.ConnCopier {
	client, err := f.getSSHClient()
	if err != nil {
		return dstconn.ClosingConnCopier{}
	}

	return client.NewConnCopier(proxyDesc)
}

func (f *ReconnSSHClient) Connect() error {
	_ = f.disconnect()
	_, err := f.connect()
	return err
}

func (f *ReconnSSHClient) Disconnect() error {
	return f.disconnect()
}

func (f *ReconnSSHClient) getSSHClient() (*dstconn.SSHClient, error) {
	f.sshClientLock.RLock()

	if f.sshClient != nil {
		f.sshClientLock.RUnlock()
		return f.sshClient, nil
	}

	f.sshClientLock.RUnlock()

	return f.connect()
}

func (f *ReconnSSHClient) connect() (*dstconn.SSHClient, error) {
	f.sshClientLock.Lock()
	defer f.sshClientLock.Unlock()

	// Check if client has been already reconnected in a competing connect call
	if f.sshClient != nil {
		return f.sshClient, nil
	}

	f.logger.Debug(f.logTag, "Trying to reconnect SSH client")

	sess, err := f.entryPoint.EntryPoint()
	if err != nil {
		return nil, err
	}

	sshClient := dstconn.NewSSHClient(sess.Opts(), f.logger)

	err = sshClient.Connect()
	if err != nil {
		return nil, err
	}

	f.logger.Debug(f.logTag, "Reconnected SSH client")
	f.sshClient = sshClient
	f.entryPointSess = sess

	return f.sshClient, nil
}

func (f *ReconnSSHClient) disconnect() error {
	f.sshClientLock.Lock()
	defer f.sshClientLock.Unlock()

	var err error

	if f.sshClient != nil {
		err = f.sshClient.Disconnect()
		f.sshClient = nil
	}

	if f.entryPointSess != nil {
		f.entryPointSess.Close()
	}

	return err
}
