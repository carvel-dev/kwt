package net

import (
	"fmt"
	"syscall"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kwt/pkg/kwt/cmd/core"
	ctldns "github.com/k14s/kwt/pkg/kwt/dns"
	ctlnet "github.com/k14s/kwt/pkg/kwt/net"
	"github.com/k14s/kwt/pkg/kwt/net/dstconn"
	"github.com/k14s/kwt/pkg/kwt/net/forwarder"
	"github.com/k14s/kwt/pkg/kwt/setgid"
	"github.com/spf13/cobra"
)

type StartOptions struct {
	depsFactory   cmdcore.DepsFactory
	configFactory cmdcore.ConfigFactory
	ui            ui.UI
	cancelSignals cmdcore.CancelSignals

	NamespaceFlags NamespaceFlags
	DNSFlags       DNSFlags
	LoggingFlags   LoggingFlags
	SSHFlags       SSHFlags

	Subnets   []string
	RemoteIPs []string
}

func NewStartOptions(
	depsFactory cmdcore.DepsFactory,
	configFactory cmdcore.ConfigFactory,
	ui ui.UI,
	cancelSignals cmdcore.CancelSignals,
) *StartOptions {
	return &StartOptions{
		depsFactory:   depsFactory,
		configFactory: configFactory,
		ui:            ui,
		cancelSignals: cancelSignals,
	}
}

func NewStartCmd(o *StartOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start",
		Aliases: []string{"s"},
		Short:   "Sets up network access",
		Example: `
  # Detect settings automatically
  sudo -E kwt net start

  # Provide predefined set of subnets to proxy
  sudo -E kwt net start --subnet 10.19.247.0/24 --subnet 10.19.248.0/24

  # Redirect all example.com and its subdomains to localhost
  sudo -E kwt net start --dns-map example.com=127.0.0.1

  # Dynamically configure DNS mappings
  sudo -E kwt net start --dns-map-exec='knctl dns-map'
`,
		RunE: func(_ *cobra.Command, _ []string) error { return o.Run() },
	}

	o.NamespaceFlags.Set(cmd)
	o.DNSFlags.SetWithPrefix(cmd, "dns")
	o.LoggingFlags.Set(cmd)
	o.SSHFlags.Set(cmd)

	cmd.Flags().StringSliceVarP(&o.Subnets, "subnet", "s", nil, "Subnet, if specified subnets will not be guessed automatically (can be specified multiple times)")
	cmd.Flags().StringSliceVar(&o.RemoteIPs, "remote-ip", nil, "Additional IP to include for subnet guessing (can be specified multiple times)")

	return cmd
}

func (o *StartOptions) Run() error {
	if syscall.Geteuid() != 0 {
		return fmt.Errorf("Command must run under sudo to change firewall settings (sudo -E kwt net start ...)")
	}

	gidInt, err := setgid.GidExec{}.SetProcessGID()
	if err != nil {
		return fmt.Errorf("Changing group id: %s", err)
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
	logTag := "StartOptions"

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

	var subnets ctlnet.Subnets

	if len(o.Subnets) > 0 {
		subnets = ctlnet.NewConfiguredSubnets(o.Subnets)
	} else {
		subnets = ctlnet.NewKubeSubnets(coreClient, o.RemoteIPs, logger)
	}

	dnsIPs := ResolvConfDNSIPs{ctldns.NewResolvConf()}
	dnsServerFactory := NewDNSServerFactory(o.DNSFlags, dnsIPs, coreClient, logger)
	forwarderFactory := forwarder.NewFactory(gidInt, logger)
	forwardingProxy := ctlnet.NewForwardingProxy(forwarderFactory, dnsServerFactory, logger)
	remotingProxy := ctlnet.NewRemotingProxy(entryPoint, subnets, dnsIPs, forwardingProxy, logger)

	o.cancelSignals.Watch(func() {
		logger.Info(logTag, "Shutting down")

		err := remotingProxy.Shutdown()
		if err != nil {
			logger.Error(logTag, "Failed shutting proxy: %s", err)
		}
	})

	return remotingProxy.Serve()
}
