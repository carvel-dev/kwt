package net

import (
	"fmt"
	"net"

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

	reconnListener := ctlnet.NewReconnListener(service, reconnSSHClient, logger)

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
