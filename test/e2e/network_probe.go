package e2e

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
	"crypto/tls"
)

type NetworkProbe struct {
	t      *testing.T
	logger Logger
}

func (a NetworkProbe) HTTPGet(url, expectedOutput, description string) {
	a.logger.Section(fmt.Sprintf("Test network accessibility to the HTTP service (%s) via '%s'", description, url), func() {
		client := &http.Client{
			Transport: &http.Transport{
				TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},

				Proxy: http.ProxyFromEnvironment,
				Dial:  (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,

				TLSHandshakeTimeout: 30 * time.Second,
				DisableKeepAlives:   true,
			},
		}

		res, err := client.Get(url)
		if err != nil {
			a.t.Fatalf("Error making HTTP request: %s", err)
		}

		defer res.Body.Close()

		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			a.t.Fatalf("Error reading HTTP request body: %s", err)
		}

		if !strings.Contains(string(bodyBytes), expectedOutput) {
			a.t.Fatalf("Expected to find output '%s' in body '%s'", expectedOutput, string(bodyBytes))
		}
	})
}

func (a NetworkProbe) RedisWriteRead(addr, storedValue string) {
	a.logger.Section(fmt.Sprintf("Test network accessibility to the TCP service (redis) via 'tcp://%s'", addr), func() {
		tcpAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(addr, "6379"))
		if err != nil {
			a.t.Fatalf("ResolveTCPAddr failed: %s", err)
		}

		{
			setConn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				a.t.Fatalf("Dial failed: %s", err)
			}

			defer setConn.Close()

			_, err = setConn.Write([]byte(fmt.Sprintf("set test-key %s\n", storedValue)))
			if err != nil {
				a.t.Fatalf("Write to server failed: %s", err)
			}

			replyOutput := make([]byte, 1024)

			_, err = setConn.Read(replyOutput)
			if err != nil {
				panic(fmt.Sprintf("Read from server failed: %s", err))
			}

			if !strings.Contains(string(replyOutput), "+OK") {
				a.t.Fatalf("Expected to find ok in body '%s'", string(replyOutput))
			}

			setConn.Close()
		}

		{
			getConn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				a.t.Fatalf("Dial failed: %s", err)
			}

			defer getConn.Close()

			_, err = getConn.Write([]byte("get test-key\n"))
			if err != nil {
				a.t.Fatalf("Write to server failed: %s", err)
			}

			replyOutput := make([]byte, 1024)

			_, err = getConn.Read(replyOutput)
			if err != nil {
				a.t.Fatalf("Read from server failed: %s", err)
			}

			if !strings.Contains(string(replyOutput), storedValue) {
				a.t.Fatalf("Expected to find output '%s' in body '%s'", storedValue, string(replyOutput))
			}

			getConn.Close()
		}
	})
}
