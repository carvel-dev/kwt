package net

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type ReconnListener struct {
	service         *KubeListenerService
	reconnSSHClient *ReconnSSHClient

	listener     net.Listener
	listenerLock sync.RWMutex
	closedCh     chan struct{}

	logger Logger
	logTag string
}

var _ net.Listener = &ReconnListener{}

func NewReconnListener(service *KubeListenerService, reconnSSHClient *ReconnSSHClient, logger Logger) *ReconnListener {
	return &ReconnListener{
		service:         service,
		reconnSSHClient: reconnSSHClient,
		closedCh:        make(chan struct{}),

		logger: logger,
		logTag: "ReconnListener",
	}
}

func (lis *ReconnListener) Accept() (net.Conn, error) {
	for {
		listener, err := lis.getListener()
		if err != nil {
			return nil, err
		}

		conn, err := listener.Accept()
		if err != nil {
			if err == io.EOF { // listener was closed
				select {
				case <-lis.closedCh:
					return nil, io.EOF
				default:
					lis.disconnect()
					continue
				}
			}
			return nil, err
		}

		return conn, nil
	}
}

func (lis *ReconnListener) Close() error {
	select {
	case <-lis.closedCh:
		// already closed
	default:
		close(lis.closedCh)
	}

	return lis.disconnect()
}

func (lis *ReconnListener) Addr() net.Addr {
	if lis.listener != nil {
		return lis.listener.Addr()
	}
	return fakeAddr{} // TODO better address
}

func (lis *ReconnListener) getListener() (net.Listener, error) {
	lis.listenerLock.RLock()

	if lis.listener != nil {
		defer lis.listenerLock.RUnlock()
		return lis.listener, nil
	}

	lis.listenerLock.RUnlock()

	return lis.connect()
}

func (lis *ReconnListener) connect() (net.Listener, error) {
	lis.listenerLock.Lock()
	defer lis.listenerLock.Unlock()

	var err error

	lis.listener, err = lis.reconnSSHClient.NewListener()
	if err != nil {
		return nil, err
	}

	lis.logger.Debug(lis.logTag, "Remotely listening on %s", lis.listener.Addr())

	_, targetPort, err := net.SplitHostPort(lis.listener.Addr().String())
	if err != nil {
		return nil, fmt.Errorf("Extracting target port: %s", err)
	}

	err = lis.service.Redirect(targetPort)
	if err != nil {
		return nil, fmt.Errorf("Redirecting service: %s", err)
	}

	return lis.listener, nil
}

func (lis *ReconnListener) disconnect() error {
	lis.listenerLock.Lock()
	defer lis.listenerLock.Unlock()

	var err error

	if lis.listener != nil {
		err = lis.listener.Close()
		lis.listener = nil
	}

	return err
}

type fakeAddr struct{}

var _ net.Addr = fakeAddr{}

func (fakeAddr) Network() string { return "" }
func (fakeAddr) String() string  { return "fake-addr" }
