package net

import (
	"net"

	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	ctlnet "github.com/carvel-dev/kwt/pkg/kwt/net"
	"github.com/carvel-dev/kwt/pkg/kwt/net/forwarder"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type ForwardOptions struct {
	depsFactory   cmdcore.DepsFactory
	ui            ui.UI
	cancelSignals cmdcore.CancelSignals

	LoggingFlags LoggingFlags

	Subnets []string
	TCPPort int
	UDPPort int
}

func NewForwardOptions(depsFactory cmdcore.DepsFactory, ui ui.UI, cancelSignals cmdcore.CancelSignals) *ForwardOptions {
	return &ForwardOptions{depsFactory: depsFactory, ui: ui, cancelSignals: cancelSignals}
}

func NewForwardCmd(o *ForwardOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "forward",
		Short:  "Forward subnet traffic to single port",
		RunE:   func(_ *cobra.Command, _ []string) error { return o.Run() },
		Hidden: true,
	}

	o.LoggingFlags.Set(cmd)

	cmd.Flags().StringSliceVarP(&o.Subnets, "subnet", "s", nil, "Subnet (can be specified multiple times)")
	cmd.Flags().IntVar(&o.TCPPort, "tcp-port", 8080, "TCP port destination")
	cmd.Flags().IntVar(&o.UDPPort, "udp-port", 1234, "UDP port destination")

	return cmd
}

func (o *ForwardOptions) Run() error {
	logger := cmdcore.NewLoggerWithDebug(o.ui, o.LoggingFlags.Debug)

	forwarderFactory := forwarder.NewFactory(0, logger)
	opts := forwarder.ForwarderOpts{DstTCPPort: o.TCPPort, DstDNSUDPPort: o.UDPPort}

	forwarder, err := forwarderFactory.NewForwarder(opts)
	if err != nil {
		return err
	}

	err = forwarder.CheckPrereqs()
	if err != nil {
		return err
	}

	subnets, err := ctlnet.NewConfiguredSubnets(o.Subnets).Subnets()
	if err != nil {
		return err
	}

	err = forwarder.Add(subnets, []net.IP{})
	if err != nil {
		return err
	}

	o.ui.PrintLinef("Forwarding TCP %s -> 127.0.0.1:%d", o.Subnets, o.TCPPort)
	o.ui.PrintLinef("Forwarding UDP %s -> 127.0.0.1:%d", o.Subnets, o.UDPPort)

	doneCh := make(chan struct{})

	o.cancelSignals.Watch(func() {
		o.ui.PrintLinef("Stopping")

		err := forwarder.Reset()
		if err != nil {
			logger.Error("ForwardOptions", "Failed resetting forwarder: %s", err)
		}

		doneCh <- struct{}{}
	})

	<-doneCh

	return nil
}
