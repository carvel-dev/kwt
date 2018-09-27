package e2e

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

const dscacheutilIPAddrPrefix = "ip_address: "

func TestNetmDNSResolution(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping test because not on darwin")
	}

	env := BuildEnv(t)
	logger := Logger{}
	kwt := Kwt{t, env.Namespace, Logger{}}
	kubectl := Kubectl{t, env.Namespace, Logger{}}
	kwtNet := NewKwtNet(kwt, t, Logger{})

	kwtNet.Start([]string{})
	defer kwtNet.End()

	guestbookAddrs := Guestbook{kwt, kubectl, t, logger}.Install()

	t1 := time.Now()
	out, err := exec.Command("dscacheutil", "-q", "host", "-a", "name", guestbookAddrs.FrontendSvcDomain).CombinedOutput()
	timeDesc := fmt.Sprintf("Time took to resolve %s", time.Now().Sub(t1))

	if err != nil {
		t.Fatalf("Expected to resolve guestbook frontend domain successfully: %s (%s) (output: %s)", err, timeDesc, out)
	}

	logger.Debugf("%s\n", timeDesc)

	foundAddr := ""
	outputLines := strings.Split(string(out), "\n")

	for _, line := range outputLines {
		if strings.HasPrefix(line, dscacheutilIPAddrPrefix) {
			foundAddr = strings.TrimPrefix(line, dscacheutilIPAddrPrefix)
			break
		}
	}

	if len(foundAddr) == 0 {
		t.Fatalf("Expected to find ip address in output: %s", out)
	}

	if foundAddr != guestbookAddrs.FrontendSvcIP {
		t.Fatalf("Expected resolved IP to match '%s' but was '%s'", guestbookAddrs.FrontendSvcIP, foundAddr)
	}
}
