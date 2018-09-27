package e2e

import (
	"fmt"
	"testing"
)

func TestNetTCPandHTTPMultiple(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kwt := Kwt{t, env.Namespace, Logger{}}
	kubectl := Kubectl{t, env.Namespace, Logger{}}

	guestbookAddrs := Guestbook{kwt, kubectl, t, logger}.Install()

	kwtNet1 := NewKwtNet(kwt, t, logger)
	kwtNet1.Start([]string{"--subnet", guestbookAddrs.FrontendSvcIP + "/32"})
	defer kwtNet1.End()

	kwtNet2 := NewKwtNet(kwt, t, logger)
	kwtNet2.StartWithoutCleanup([]string{"--subnet", guestbookAddrs.RedisSvcIP + "/32"})
	defer kwtNet2.EndWithoutCleanup()

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
