package e2e

import (
	"fmt"
	"regexp"
	"testing"
)

var (
	tcpProxyStartLine    = `info: TCPProxy: Started proxy on`
	tcpProxyStartedLine  = `info: TCPProxy: Started \d+\.\d+\.\d+\.\d+:\d+\-\>`
	tcpProxyFinishedLine = `info: TCPProxy: Finished \d+\.\d+\.\d+\.\d+:\d+\-\>`
)

func TestNetTCPandHTTPSingle(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kwt := Kwt{t, env.Namespace, Logger{}}
	kubectl := Kubectl{t, env.Namespace, Logger{}}
	kwtNet := NewKwtNet(kwt, t, Logger{})

	kwtNet.Start([]string{})
	defer kwtNet.End()

	guestbookAddrs := Guestbook{kwt, kubectl, t, logger}.Install()
	netProbe := NetworkProbe{t, logger}

	beforeOutput := kwtNet.CollectedOutput()

	for _, url := range []string{
		fmt.Sprintf("http://%s", guestbookAddrs.FrontendSvcIP),
		fmt.Sprintf("http://%s", guestbookAddrs.FrontendSvcDomain),
	} {
		netProbe.HTTPGet(url, "Guestbook", "guestbook")
	}

	for i, addr := range []string{guestbookAddrs.RedisSvcIP, guestbookAddrs.RedisSvcDomain} {
		netProbe.RedisWriteRead(addr, fmt.Sprintf("value%d", i))
	}

	afterOutput := kwtNet.CollectedOutput()

	if !regexp.MustCompile(tcpProxyStartLine).MatchString(beforeOutput) {
		t.Fatalf("Expected to find tcp proxy start info line in output '%s'", afterOutput)
	}

	logger.Section("Check that command output shows HTTP requests made", func() {
		startRegexp := regexp.MustCompile(tcpProxyStartedLine + regexp.QuoteMeta(guestbookAddrs.FrontendSvcIP) + ":80")
		finishedRegexp := regexp.MustCompile(tcpProxyFinishedLine + regexp.QuoteMeta(guestbookAddrs.FrontendSvcIP) + ":80")
		checkTCPProxyOutput(t, startRegexp, finishedRegexp, beforeOutput, afterOutput)
	})

	logger.Section("Check that command output shows TCP requests made", func() {
		startRegexp := regexp.MustCompile(tcpProxyStartedLine + regexp.QuoteMeta(guestbookAddrs.RedisSvcIP) + ":6379")
		finishedRegexp := regexp.MustCompile(tcpProxyFinishedLine + regexp.QuoteMeta(guestbookAddrs.RedisSvcIP) + ":6379")
		checkTCPProxyOutput(t, startRegexp, finishedRegexp, beforeOutput, afterOutput)
	})
}

func checkTCPProxyOutput(t *testing.T, startedLine, finishedLine *regexp.Regexp, beforeOutput, afterOutput string) {
	if startedLine.MatchString(beforeOutput) {
		t.Fatalf("Expected to NOT find tcp proxy started info line in output '%s'", beforeOutput)
	}
	if finishedLine.MatchString(beforeOutput) {
		t.Fatalf("Expected to NOT find tcp proxy finished info line in output '%s'", beforeOutput)
	}

	if !startedLine.MatchString(afterOutput) {
		t.Fatalf("Expected to find tcp proxy started info line in output '%s'", afterOutput)
	}
	if !finishedLine.MatchString(afterOutput) {
		t.Fatalf("Expected to find tcp proxy finished info line in output '%s'", afterOutput)
	}
}

/*

Example output:

03:55:10PM: info: mdns.CustomHandler: A:frontend.default.svc.cluster.local.,AAAA:frontend.default.svc.cluster.local.: Answering (5.891237ms)
03:55:11PM: info: mdns.CustomHandler: AAAA:frontend.default.svc.cluster.local.: Answering (2.481284ms)
03:55:13PM: info: TCPProxy: Received 10.80.130.76:63587
03:55:13PM: info: TCPProxy: Started 10.80.130.76:63587->10.98.108.253:80
03:55:13PM: info: TCPProxy: Finished 10.80.130.76:63587->10.98.108.253:80 (1.638324ms/61.823787ms)
03:55:30PM: info: dns.ForwardHandler: TXT:time-osx.g.aaplimg.com.: Answering via=10.87.8.10:53 (721.326µs)

*/
