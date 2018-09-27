package e2e

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

type Guestbook struct {
	kwt     Kwt
	kubectl Kubectl
	t       *testing.T
	logger  Logger
}

type GuestbookAddrs struct {
	FrontendSvcIP, FrontendSvcDomain string
	RedisSvcIP, RedisSvcDomain       string
}

func (a GuestbookAddrs) Complete() bool {
	return len(a.FrontendSvcIP) > 0 && len(a.FrontendSvcDomain) > 0 &&
		len(a.RedisSvcIP) > 0 && len(a.RedisSvcDomain) > 0
}

func (g Guestbook) Install() GuestbookAddrs {
	g.logger.Section("Install kubernetes guestbook", func() {
		cwdPath, err := os.Getwd()
		if err != nil {
			g.t.Fatalf("Expected not to fail getting current directory: %s", err)
		}

		guestbookBytes, err := ioutil.ReadFile(filepath.Join(cwdPath, "assets", "guestbook-all-in-one.yml"))
		if err != nil {
			g.t.Fatalf("Expected not to fail reading guestbook file: %s", err)
		}

		g.kubectl.RunWithOpts([]string{"apply", "-f", "-"}, RunOpts{StdinReader: bytes.NewReader(guestbookBytes)})

		if len(os.Getenv("KWT_GUESTBOOK_SKIP_WAIT")) == 0 {
			time.Sleep(20 * time.Second) // TODO remove
		}
	})

	addrs := GuestbookAddrs{}

	g.logger.Section("Wait guestbook services to be available", func() {
		timeoutCh := time.After(2 * time.Minute)

		for {
			out := g.kwt.Run([]string{"net", "svc", "--json"})
			resp := uitest.JSONUIFromBytes(g.t, []byte(out))

			for _, row := range resp.Tables[0].Rows {
				if row["name"] == "frontend" {
					addrs.FrontendSvcIP = row["cluster_ip"]
					addrs.FrontendSvcDomain = row["internal_dns"]
				}
				if row["name"] == "redis-master" {
					addrs.RedisSvcIP = row["cluster_ip"]
					addrs.RedisSvcDomain = row["internal_dns"]
				}
			}

			if addrs.Complete() {
				break
			}

			select {
			case <-timeoutCh:
				g.t.Fatalf("Timed out waiting for guestbook svc cluster IP and DNS to be seen in output: %s", out)
			default:
				// continue with waiting
			}

			time.Sleep(1 * time.Second)
		}
	})

	return addrs
}
