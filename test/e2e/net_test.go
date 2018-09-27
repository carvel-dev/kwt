package e2e

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestNetTCPandHTTP(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kwt := Kwt{t, env.Namespace, Logger{}}
	kubectl := Kubectl{t, env.Namespace, Logger{}}
	kwtNet := NewKwtNet(kwt, t, Logger{})

	kwtNet.Start()
	defer kwtNet.End()

	guestbookAddrs := Guestbook{kwt, kubectl, t, logger}.Install()

	for _, url := range []string{
		fmt.Sprintf("http://%s", guestbookAddrs.FrontendSvcIP),
		fmt.Sprintf("http://%s", guestbookAddrs.FrontendSvcDomain),
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

	for i, addr := range []string{guestbookAddrs.RedisSvcIP, guestbookAddrs.RedisSvcDomain} {
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
