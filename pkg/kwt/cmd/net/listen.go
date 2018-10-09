package net

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	ctlnet "github.com/cppforlife/kwt/pkg/kwt/net"
	"github.com/cppforlife/kwt/pkg/kwt/net/dstconn"
	"github.com/cppforlife/kwt/pkg/kwt/net/forwarder"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

type ListenOptions struct {
	depsFactory   cmdcore.DepsFactory
	configFactory cmdcore.ConfigFactory
	ui            ui.UI
	cancelSignals cmdcore.CancelSignals

	NamespaceFlags cmdcore.NamespaceFlags

	Service     string
	ServiceType string

	RemoteAddr string
	LocalAddr  string

	SSHFlags     SSHFlags
	LoggingFlags LoggingFlags
}

func NewListenOptions(
	depsFactory cmdcore.DepsFactory,
	configFactory cmdcore.ConfigFactory,
	ui ui.UI,
	cancelSignals cmdcore.CancelSignals,
) *ListenOptions {
	return &ListenOptions{
		depsFactory:   depsFactory,
		configFactory: configFactory,
		ui:            ui,
		cancelSignals: cancelSignals,
	}
}

func NewListenCmd(o *ListenOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "listen",
		Aliases: []string{"l"},
		Short:   "Redirect incoming service traffic to a local port",
		Example: `
	# Create service 'svc1' and forward its port 80 to localhost:80
	kwt net listen --service svc1

	# Create service 'svc1' and forward its port 8080 to localhost:8081
	kwt net listen -r 8080 -l localhost:8081 --service svc1

	# TODO Create service and forward it to local listening process
	kwt net l --service svc1 --local-process foo
`,
		RunE: func(_ *cobra.Command, _ []string) error { return o.Run() },
	}

	// TODO how to connect across multiple namespaces
	o.NamespaceFlags.Set(cmd, flagsFactory)

	cmd.Flags().StringVarP(&o.Service, "service", "s", "", "Service to create or update for incoming traffic")
	cmd.Flags().StringVar(&o.ServiceType, "service-type", "ClusterIP", "Service type to set if creating service")

	cmd.Flags().StringVarP(&o.RemoteAddr, "remote", "r", "80", "Remote address (example: 80)")
	cmd.Flags().StringVarP(&o.LocalAddr, "local", "l", "localhost:80", "Local address (example: 80, localhost:80)")

	o.SSHFlags.Set(cmd)
	o.LoggingFlags.Set(cmd)

	return cmd
}

func (o *ListenOptions) Run() error {
	if len(o.Service) == 0 {
		return fmt.Errorf("Expected non-empty service name")
	}

	localAddr, err := net.ResolveTCPAddr("tcp", o.LocalAddr)
	if err != nil {
		return fmt.Errorf("Resolving local addr: %s", err)
	}

	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	restConfig, err := o.configFactory.RESTConfig()
	if err != nil {
		return err
	}

	logger := cmdcore.NewLoggerWithDebug(o.ui, o.LoggingFlags.Debug)
	logTag := "ListenOptions"

	var entryPoint ctlnet.EntryPoint

	if len(o.SSHFlags.PrivateKey) > 0 {
		entryPoint = ctlnet.NewSSHEntryPoint(dstconn.SSHClientConnOpts{
			User:          o.SSHFlags.User,
			Host:          o.SSHFlags.Host,
			PrivateKeyPEM: o.SSHFlags.PrivateKey,
		})
	} else {
		entryPoint = ctlnet.NewKubeEntryPoint(coreClient, restConfig, o.NamespaceFlags.Name, logger)
	}

	reconnSSHClient := ctlnet.NewReconnSSHClient(entryPoint, logger)

	err = reconnSSHClient.Connect()
	if err != nil {
		return err
	}

	defer reconnSSHClient.Disconnect()

	service := ctlnet.NewKubeListenerService(
		o.Service, corev1.ServiceType(o.ServiceType), o.NamespaceFlags.Name, o.RemoteAddr, coreClient, logger)

	err = service.Snapshot()
	if err != nil {
		return fmt.Errorf("Snapshotting service: %s", err)
	}

	defer func() {
		err := service.Revert()
		if err != nil {
			logger.Error(logTag, "Failed reverting service redirection: %s", err)
		}
	}()

	reconnListener := NewReconnListener(service, reconnSSHClient, logger)

	defer reconnListener.Close()

	o.cancelSignals.Watch(func() {
		logger.Info(logTag, "Shutting down")
		reconnListener.Close() // TODO use proxy.Shutdown()?
	})

	resolver := forwarder.NewStaticResolver(localAddr.IP, localAddr.Port)
	proxy := ctlnet.NewTCPProxy(resolver, dstconn.NewLocal(logger), logger)
	startedCh := make(chan struct{})

	go func() {
		<-startedCh
		logger.Info(logTag, "Forwarding %s->%s", o.RemoteAddr, o.LocalAddr)
		logger.Info(logTag, "Ready!")
	}()

	return proxy.ServeListener(reconnListener, startedCh)
}

type ReconnListener struct {
	service         *ctlnet.KubeListenerService
	reconnSSHClient *ctlnet.ReconnSSHClient

	listener     net.Listener
	listenerLock sync.RWMutex
	closedCh     chan struct{}

	logger ctlnet.Logger
	logTag string
}

var _ net.Listener = &ReconnListener{}

func NewReconnListener(service *ctlnet.KubeListenerService, reconnSSHClient *ctlnet.ReconnSSHClient, logger ctlnet.Logger) *ReconnListener {
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
	return dummyAddr{} // TODO better address
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

type dummyAddr struct{}

var _ net.Addr = dummyAddr{}

func (dummyAddr) Network() string { return "" }
func (dummyAddr) String() string  { return "dummy-addr" }
