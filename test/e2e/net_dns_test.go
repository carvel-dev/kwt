package e2e

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"testing"
)

var (
	dnsServerStartLine = `info: dns.Server: Started DNS server on`
	dnsCustomHandlerLine = `info: dns.CustomHandler: A:%s\.: Answering rcode=0`
)

type ResolverDialer struct {
	Name string
	Func func(ctx context.Context, network, address string) (net.Conn, error)
}

func TestNetDNSResolution(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kwt := Kwt{t, env.Namespace, Logger{}}
	kubectl := Kubectl{t, env.Namespace, Logger{}}
	kwtNet := NewKwtNet(kwt, t, Logger{})

	kwtNet.Start([]string{})
	defer kwtNet.End()

	guestbookAddrs := Guestbook{kwt, kubectl, t, logger}.Install()

	tcpDialer := ResolverDialer{
		Name: "TCP",
		Func: func(ctx context.Context, network, address string) (net.Conn, error) {
			addr, err := net.ResolveTCPAddr("tcp", address)
			if err != nil {
				return nil, err
			}
			return net.DialTCP("tcp", nil, addr)
		},
	}

	udpDialer := ResolverDialer{
		Name: "UDP",
		Func: func(ctx context.Context, network, address string) (net.Conn, error) {
			addr, err := net.ResolveUDPAddr("udp", address)
			if err != nil {
				return nil, err
			}
			return net.DialUDP("udp", nil, addr)
		},
	}

	beforeOutput := kwtNet.CollectedOutput()

	for _, dialer := range []ResolverDialer{tcpDialer, udpDialer} {
		dialer := dialer

		logger.Section(fmt.Sprintf("Test DNS resolution over %s", dialer.Name), func() {
			dnsResolver := net.Resolver{ // Not going to mdns
				PreferGo:     true,
				StrictErrors: true,
				Dial:         dialer.Func,
			}

			addrs, err := dnsResolver.LookupIPAddr(context.TODO(), guestbookAddrs.FrontendSvcDomain)
			if err != nil {
				t.Fatalf("Expected DNS resolution to succeed: %s", err)
			}

			if len(addrs) != 1 {
				t.Fatalf("Expected DNS resolution to result in one address, but was '%#v'", addrs)
			}

			if addrs[0].String() != guestbookAddrs.FrontendSvcIP {
				t.Fatalf("Expected DNS resolution to result in one specific address '%s', but was '%s'",
					guestbookAddrs.FrontendSvcIP, addrs[0].String())
			}
		})
	}

	afterOutput := kwtNet.CollectedOutput()

	if !regexp.MustCompile(dnsServerStartLine).MatchString(beforeOutput) {
		t.Fatalf("Expected to find dns start info line in output '%s'", afterOutput)
	}

	logger.Section("Check that command output shows DNS custom handler requests answered", func() {
		lineRegexp := regexp.MustCompile(fmt.Sprintf(dnsCustomHandlerLine, regexp.QuoteMeta(guestbookAddrs.FrontendSvcDomain)))
		checkDNSCustomHandlerOutput(t, lineRegexp, beforeOutput, afterOutput)
	})
}

func checkDNSCustomHandlerOutput(t *testing.T, line *regexp.Regexp, beforeOutput, afterOutput string) {
	if line.MatchString(beforeOutput) {
		t.Fatalf("Expected to NOT find dns.CustomHandler info line in output '%s'", beforeOutput)
	}
	if !line.MatchString(afterOutput) {
		t.Fatalf("Expected to find dns.CustomHandler info line in output '%s'", afterOutput)
	}
}

/*

Example output:

04:50:10PM: info: dns.CustomHandler: A:frontend.default.svc.cluster.local.: Answering rcode=0 (4.921799ms)
04:51:11PM: info: dns.CustomHandler: AAAA:frontend.default.svc.cluster.local.: Answering rcode=0 (33.162µs)

*/
