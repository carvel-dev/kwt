package net

import (
	"net"
	"sync"

	"github.com/cppforlife/kwt/pkg/kwt/net/dstconn"
)

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
		client, err := f.reconnectIfNecessary(err)
		if err != nil {
			return nil, err
		}
		return client.NewConn(ip, port)
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

func (f *ReconnSSHClient) NewListener() (net.Listener, error) {
	client, err := f.getSSHClient()
	if err != nil {
		return nil, err
	}

	lis, err := client.NewListener()
	if err != nil {
		client, err := f.reconnectIfNecessary(err)
		if err != nil {
			return nil, err
		}
		return client.NewListener()
	}

	return lis, nil
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
		defer f.sshClientLock.RUnlock()
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

func (f *ReconnSSHClient) reconnectIfNecessary(err error) (*dstconn.SSHClient, error) {
	_, needsReconnect := err.(dstconn.ConnectionBrokenErr)

	f.logger.Debug(f.logTag, "Received err: %s (needsReconnect: %t)", err, needsReconnect)

	if needsReconnect {
		f.disconnect()
		return f.connect()
	}

	return nil, err
}
