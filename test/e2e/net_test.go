package e2e

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestNet(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kwt := Kwt{t, env.Namespace, Logger{}}
	kubectl := Kubectl{t, env.Namespace, Logger{}}

	cleanUp := func() {
		logger.Section("Clean up net access endpoint", func() {
			kwt.RunWithOpts([]string{"net", "clean-up"}, RunOpts{AllowError: true, NoNamespace: true})
		})
	}

	cleanUp()
	defer cleanUp()

	cancelCh := make(chan struct{})
	doneCh := make(chan struct{})
	collectedOutput := &LogsWriter{}

	logger.Section("Starting net start in background", func() {
		go func() {
			kwt.RunWithOpts([]string{"net", "start", "--tty"}, RunOpts{StdoutWriter: collectedOutput, CancelCh: cancelCh, NoNamespace: true})
			doneCh <- struct{}{}
		}()
	})

	// Clean up on failure
	defer func() {
		logger.Section("Terminating net start tailing", func() {
			cancelCh <- struct{}{}
			<-doneCh
		})
	}()

	logger.Section("Wait for forwarding to be ready", func() {
		timeoutCh := time.After(2 * time.Minute)
		const expectedOutput = "Ready!"

		for {
			if strings.Contains(collectedOutput.Current(), expectedOutput) {
				break
			}

			select {
			case <-timeoutCh:
				t.Fatalf("Timed out waiting for '%s' to be seen in output '%s'", expectedOutput, collectedOutput.Current())
			default:
				// continue with waiting
			}

			time.Sleep(1 * time.Second)
		}
	})

	logger.Section("Install kubernetes guestbook", func() {
		cwdPath, err := os.Getwd()
		if err != nil {
			t.Fatalf("Expected not to fail getting current directory: %s", err)
		}

		guestbookBytes, err := ioutil.ReadFile(filepath.Join(cwdPath, "assets", "guestbook-all-in-one.yml"))
		if err != nil {
			t.Fatalf("Expected not to fail reading guestbook file: %s", err)
		}

		kubectl.RunWithOpts([]string{"apply", "-f", "-"}, RunOpts{StdinReader: bytes.NewReader(guestbookBytes)})

		time.Sleep(20 * time.Second) // TODO remove
	})

	var guestbookFrontendSvcIP, guestbookFrontendSvcDomain string
	var guestbookRedisSvcIP, guestbookRedisSvcDomain string

	logger.Section("Wait guestbook services to be available", func() {
		timeoutCh := time.After(2 * time.Minute)

		for {
			out := kwt.Run([]string{"net", "svc", "--json"})
			resp := uitest.JSONUIFromBytes(t, []byte(out))

			for _, row := range resp.Tables[0].Rows {
				if row["name"] == "frontend" {
					guestbookFrontendSvcIP = row["cluster_ip"]
					guestbookFrontendSvcDomain = row["internal_dns"]
				}
				if row["name"] == "redis-master" {
					guestbookRedisSvcIP = row["cluster_ip"]
					guestbookRedisSvcDomain = row["internal_dns"]
				}
			}

			if len(guestbookFrontendSvcIP) > 0 && len(guestbookFrontendSvcDomain) > 0 {
				break
			}

			select {
			case <-timeoutCh:
				t.Fatalf("Timed out waiting for guestbook svc cluster IP and DNS to be seen in output: %s", out)
			default:
				// continue with waiting
			}

			time.Sleep(1 * time.Second)
		}
	})

	for _, url := range []string{
		fmt.Sprintf("http://%s", guestbookFrontendSvcIP),
		fmt.Sprintf("http://%s", guestbookFrontendSvcDomain),
	} {
		url := url

		logger.Section(fmt.Sprintf("Test network accessibility to the HTTP service (guestbook) via '%s'", url), func() {
			const expectedOutput = "Guestbook"

			res, err := http.Get(url)
			if err != nil {
				t.Fatalf("Error making HTTP request: %s", err)
			}

			defer res.Body.Close()

			bodyBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Error reading HTTP request body: %s", err)
			}

			if !strings.Contains(string(bodyBytes), expectedOutput) {
				t.Fatalf("Expected to find output '%s' in body '%s'", expectedOutput, string(bodyBytes))
			}
		})
	}

	for i, addr := range []string{guestbookRedisSvcIP, guestbookRedisSvcDomain} {
		addr := addr
		expectedOutput := fmt.Sprintf("value%d", i)

		logger.Section(fmt.Sprintf("Test network accessibility to the TCP service (redis) via 'tcp://%s'", addr), func() {
			tcpAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(addr, "6379"))
			if err != nil {
				t.Fatalf("ResolveTCPAddr failed: %s", err)
			}

			{
				setConn, err := net.DialTCP("tcp", nil, tcpAddr)
				if err != nil {
					t.Fatalf("Dial failed: %s", err)
				}

				defer setConn.Close()

				_, err = setConn.Write([]byte(fmt.Sprintf("set test-key %s\n", expectedOutput)))
				if err != nil {
					t.Fatalf("Write to server failed: %s", err)
				}

				replyOutput := make([]byte, 1024)

				_, err = setConn.Read(replyOutput)
				if err != nil {
					panic(fmt.Sprintf("Read from server failed: %s", err))
				}

				if !strings.Contains(string(replyOutput), "+OK") {
					t.Fatalf("Expected to find ok in body '%s'", string(replyOutput))
				}

				setConn.Close()
			}

			{
				getConn, err := net.DialTCP("tcp", nil, tcpAddr)
				if err != nil {
					t.Fatalf("Dial failed: %s", err)
				}

				defer getConn.Close()

				_, err = getConn.Write([]byte("get test-key\n"))
				if err != nil {
					t.Fatalf("Write to server failed: %s", err)
				}

				replyOutput := make([]byte, 1024)

				_, err = getConn.Read(replyOutput)
				if err != nil {
					t.Fatalf("Read from server failed: %s", err)
				}

				if !strings.Contains(string(replyOutput), expectedOutput) {
					t.Fatalf("Expected to find output '%s' in body '%s'", expectedOutput, string(replyOutput))
				}

				getConn.Close()
			}
		})
	}
}

type LogsWriter struct {
	lock   sync.RWMutex
	output []byte
}

var _ io.Writer = &LogsWriter{}

func (w *LogsWriter) Write(p []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.output = append(w.output, p...)
	return len(p), nil
}

func (w *LogsWriter) Current() string {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return string(w.output)
}
