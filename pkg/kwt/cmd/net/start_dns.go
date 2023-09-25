package net

import (
	"fmt"
	"net"
	"strconv"
	"syscall"

	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	ctldns "github.com/carvel-dev/kwt/pkg/kwt/dns"
	ctlfwd "github.com/carvel-dev/kwt/pkg/kwt/net/forwarder"
	"github.com/carvel-dev/kwt/pkg/kwt/setgid"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type StartDNSOptions struct {
	depsFactory   cmdcore.DepsFactory
	ui            ui.UI
	cancelSignals cmdcore.CancelSignals

	DNSFlags     DNSFlags
	LoggingFlags LoggingFlags
}

func NewStartDNSOptions(depsFactory cmdcore.DepsFactory, ui ui.UI, cancelSignals cmdcore.CancelSignals) *StartDNSOptions {
	return &StartDNSOptions{depsFactory: depsFactory, ui: ui, cancelSignals: cancelSignals}
}

func NewStartDNSCmd(o *StartDNSOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-dns",
		Short: "Start DNS server and redirect system DNS resolution to it",
		Example: `
  # Redirect all example.com and its subdomains to localhost
  sudo -E kwt net start-dns --map example.com=127.0.0.1

  # Dynamically configure DNS mappings
  sudo -E kwt net start-dns --map-exec='knctl dns-map'
`,
		RunE:   func(_ *cobra.Command, _ []string) error { return o.Run() },
		Hidden: true,
	}
	o.DNSFlags.Set(cmd)
	o.LoggingFlags.Set(cmd)
	return cmd
}

func (o *StartDNSOptions) Run() error {
	if syscall.Geteuid() != 0 {
		return fmt.Errorf("Command must run under sudo to change firewall settings (sudo -E kwt net start-dns ...)")
	}

	gidInt, err := setgid.GidExec{}.SetProcessGID()
	if err != nil {
		return fmt.Errorf("Changing group id: %s", err)
	}

	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	logger := cmdcore.NewLoggerWithDebug(o.ui, o.LoggingFlags.Debug)
	logTag := "StartDNSOptions"

	dnsIPs := ResolvConfDNSIPs{ctldns.NewResolvConf()}
	dnsServerFactory := NewDNSServerFactory(o.DNSFlags, dnsIPs, coreClient, logger)
	forwarderFactory := ctlfwd.NewFactory(gidInt, logger)

	dnsServer, err := dnsServerFactory.NewDNSServer(nil)
	if err != nil {
		return err
	}

	shutdownCh := make(chan error)

	o.cancelSignals.Watch(func() {
		shutdownCh <- nil
	})

	dnsServerErrCh := make(chan error)
	dnsServerStartedCh := make(chan struct{})

	forwarder := ctlfwd.NewLocking()
	forwarderErrCh := make(chan error)

	go func() {
		dnsServerErrCh <- dnsServer.Serve(dnsServerStartedCh)
	}()

	go func() {
		<-dnsServerStartedCh

		dnsTCPPort, err := o.portFromAddr(dnsServer.TCPAddr())
		if err != nil {
			forwarderErrCh <- err
			return
		}

		dnsUDPPort, err := o.portFromAddr(dnsServer.UDPAddr())
		if err != nil {
			forwarderErrCh <- err
			return
		}

		opts := ctlfwd.ForwarderOpts{
			DstTCPPort:    1234,
			DstDNSTCPPort: dnsTCPPort,
			DstDNSUDPPort: dnsUDPPort,
		}

		actualForwarder, err := forwarderFactory.NewForwarder(opts)
		if err != nil {
			forwarderErrCh <- err
			return
		}

		forwarder.SetForwarder(actualForwarder)

		ips, err := dnsIPs.DNSIPs()
		if err != nil {
			forwarderErrCh <- err
			return
		}

		err = forwarder.Add([]net.IPNet{}, ips)
		if err != nil {
			forwarderErrCh <- err
			return
		}

		dnsServerFactory.NewDNSOSCache().Flush()

		logger.Info(logTag, "Ready!")
	}()

	errCh := make(chan error)

	go func() {
		select {
		case <-shutdownCh:
			errCh <- nil
		case err := <-dnsServerErrCh:
			errCh <- err
		case err := <-forwarderErrCh:
			errCh <- err
		}
	}()

	origErr := <-errCh

	logger.Info(logTag, "Shutting down")

	err = forwarder.Reset()
	if err != nil {
		logger.Error(logTag, "Failed resetting forwarder: %s", err)
	}

	err = dnsServer.Shutdown()
	if err != nil {
		logger.Error(logTag, "Failed shutting down DNS server: %s", err)
	}

	return origErr
}

func (*StartDNSOptions) portFromAddr(addr net.Addr) (int, error) {
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
