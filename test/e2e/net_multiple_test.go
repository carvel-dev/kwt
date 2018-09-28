package e2e

import (
	"fmt"
	"strings"
	"testing"
)

var (
	forwardLine = `info: ForwardingProxy: Forwarding subnets: `
)

func TestNetTCPandHTTPMultiple(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kwt := Kwt{t, env.Namespace, Logger{}}
	kubectl := Kubectl{t, env.Namespace, Logger{}}

	guestbookAddrs := Guestbook{kwt, kubectl, t, logger}.Install()
	subnet1 := guestbookAddrs.FrontendSvcIP + "/32"
	subnet2 := guestbookAddrs.RedisSvcIP + "/32"

	kwtNet1 := NewKwtNet(kwt, t, logger)
	kwtNet1.Start([]string{"--subnet", subnet1})
	defer kwtNet1.End()

	output := kwtNet1.CollectedOutput()
	if !strings.Contains(output, forwardLine+subnet1) {
		t.Fatalf("Expected to find forward line info line in output '%s'", output)
	}

	kwtNet2 := NewKwtNet(kwt, t, logger)
	kwtNet2.StartWithoutCleanup([]string{"--subnet", subnet2})
	defer kwtNet2.EndWithoutCleanup()

	output = kwtNet2.CollectedOutput()
	if !strings.Contains(output, forwardLine+subnet2) {
		t.Fatalf("Expected to find forward line info line in output '%s'", output)
	}

	// TODO if multiple sessions are started, DNS isnt properly split
	// but rather goes to whichever session was started first
	netProbe := NetworkProbe{t, logger}

	for _, url := range []string{
		fmt.Sprintf("http://%s", guestbookAddrs.FrontendSvcIP),
		fmt.Sprintf("http://%s", guestbookAddrs.FrontendSvcDomain),
	} {
		netProbe.HTTPGet(url, "Guestbook", "guestbook")
	}

	for i, addr := range []string{guestbookAddrs.RedisSvcIP, guestbookAddrs.RedisSvcDomain} {
		netProbe.RedisWriteRead(addr, fmt.Sprintf("value%d", i))
	}
}
