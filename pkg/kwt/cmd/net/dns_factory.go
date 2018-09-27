package net

import (
	"fmt"
	"net"
	"strings"

	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	ctldns "github.com/cppforlife/kwt/pkg/kwt/dns"
	ctlkubedns "github.com/cppforlife/kwt/pkg/kwt/kubedns"
	ctlmdns "github.com/cppforlife/kwt/pkg/kwt/mdns"
	ctlnet "github.com/cppforlife/kwt/pkg/kwt/net"
	"github.com/cppforlife/kwt/pkg/kwt/net/dstconn"
	"k8s.io/client-go/kubernetes"
)

type DNSServerFactory struct {
	dnsFlags           DNSFlags
	defaultRecursorIPs ctlnet.DNSIPs
	coreClient         kubernetes.Interface
	logger             cmdcore.Logger
}

var _ ctlnet.DNSServerFactory = DNSServerFactory{}

func NewDNSServerFactory(dnsFlags DNSFlags, defaultRecursorIPs ctlnet.DNSIPs, coreClient kubernetes.Interface, logger cmdcore.Logger) DNSServerFactory {
	return DNSServerFactory{dnsFlags, defaultRecursorIPs, coreClient, logger}
}

func (f DNSServerFactory) NewDNSServer(dstConnFactory dstconn.Factory) (ctlnet.DNSServer, error) {
	kubeIPResolver := ctlkubedns.NewKubeDNSIPResolver(f.coreClient)

	opts, err := f.buildServerOpts(kubeIPResolver)
	if err != nil {
		return nil, err
	}

	server, err := ctldns.NewFactory().Build(opts, f.logger)
	if err != nil {
		return nil, fmt.Errorf("Building server: %s", err)
	}

	if f.dnsFlags.MDNS {
		mdnsServer := ctlmdns.NewFactory().Build(kubeIPResolver, f.logger)
		return CombinedDNSServer{server, mdnsServer}, nil
	}

	return server, nil
}

func (f DNSServerFactory) buildServerOpts(kubeIPResolver ctldns.IPResolver) (ctldns.BuildOpts, error) {
	opts := ctldns.BuildOpts{
		ListenAddrs:   []string{"localhost:0"},
		RecursorAddrs: f.dnsFlags.Recursors,

		Domains: map[string]ctldns.IPResolver{
			// Add mdns domain to regular resolver since some programs
			// may just use /etc/resolv.conf for DNS resolution on OS X (eg dig)
			// instead of relying on standard OS X resolution libraries
			ctlmdns.Domain: kubeIPResolver,
		},
	}

	if len(opts.RecursorAddrs) == 0 {
		if f.defaultRecursorIPs != nil {
			ips, err := f.defaultRecursorIPs.DNSIPs()
			if err != nil {
				return ctldns.BuildOpts{}, fmt.Errorf("Determining default DNS recursor IPs")
			}
			for _, ip := range ips {
				opts.RecursorAddrs = append(opts.RecursorAddrs, net.JoinHostPort(ip.String(), "53"))
			}
		}

		if len(opts.RecursorAddrs) == 0 {
			opts.RecursorAddrs = []string{"8.8.8.8:53"}
		}
	}

	for _, val := range f.dnsFlags.Map {
		pieces := strings.SplitN(val, "=", 2)
		if len(pieces) != 2 {
			return ctldns.BuildOpts{}, fmt.Errorf("Expected domain to IP mapping to be in format 'domain=ip' but was '%s'", val)
		}

		ip := net.ParseIP(pieces[1])
		if ip == nil {
			return ctldns.BuildOpts{}, fmt.Errorf("Expected domain to IP mapping to have valid IP '%s'", val)
		}

		opts.Domains[pieces[0]] = ctldns.NewStaticIPResolver(ip)
	}

	return opts, nil
}

type ResolvConfDNSIPs struct {
	ctldns.ResolvConf
}

var _ ctlnet.DNSIPs = ResolvConfDNSIPs{}

func (c ResolvConfDNSIPs) DNSIPs() ([]net.IP, error) {
	return c.ResolvConf.Nameservers()
}

type CombinedDNSServer struct {
	dnsServer  ctldns.Server
	mdnsServer *ctlmdns.Server
}

var _ ctlnet.DNSServer = CombinedDNSServer{}

func (s CombinedDNSServer) Serve(startedCh chan struct{}) error {
	internalStartedCh := make(chan struct{})
	errCh := make(chan error)

	go func() { errCh <- s.dnsServer.Serve(internalStartedCh) }()
	go func() { errCh <- s.mdnsServer.Serve(internalStartedCh) }()

	go func() {
		<-internalStartedCh
		<-internalStartedCh
		startedCh <- struct{}{}
	}()

	var errs []error
	errs = append(errs, <-errCh)
	errs = append(errs, <-errCh)
	return s.firstErrOrNil(errs)
}

func (s CombinedDNSServer) TCPAddr() net.Addr { return s.dnsServer.TCPAddr() }
func (s CombinedDNSServer) UDPAddr() net.Addr { return s.dnsServer.UDPAddr() }

func (s CombinedDNSServer) Shutdown() error {
	var errs []error
	errs = append(errs, s.dnsServer.Shutdown())
	errs = append(errs, s.mdnsServer.Shutdown())
	return s.firstErrOrNil(errs)
}

func (s CombinedDNSServer) firstErrOrNil(errs []error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
