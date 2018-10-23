package e2e

import (
	"net/http"
	"testing"
	"time"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestNetListen(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kwt := Kwt{t, env.Namespace, Logger{}}
	kwtNetStart := NewKwtNet(kwt, t, Logger{})
	kwtNetListen := NewKwtNet(kwt, t, Logger{})

	kwtNetStart.Start([]string{})
	defer kwtNetStart.End()

	localAddr := "localhost:8080"
	svcName := "kwt-listen-web"

	kwtNetListen.Listen([]string{"--local", localAddr, "--service", svcName})
	defer kwtNetListen.End()

	webAddr := ""
	expectedOutput := "test-reply"

	logger.Section("Wait service to be available", func() {
		timeoutCh := time.After(2 * time.Minute)

		for {
			out := kwt.Run([]string{"net", "svc", "--json"})
			resp := uitest.JSONUIFromBytes(t, []byte(out))

			for _, row := range resp.Tables[0].Rows {
				if row["name"] == svcName {
					webAddr = row["internal_dns"]
				}
			}

			if len(webAddr) > 0 {
				break
			}

			select {
			case <-timeoutCh:
				t.Fatalf("Timed out waiting for web svc cluster DNS to be seen in output: %s", out)
			default:
				// continue with waiting
			}

			time.Sleep(1 * time.Second)
		}
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expectedOutput))
	})

	server := &http.Server{Addr: localAddr, Handler: handler}
	defer server.Close()

	go server.ListenAndServe()

	netProbe := NetworkProbe{t, logger}
	netProbe.HTTPGet("http://"+webAddr, expectedOutput, "web")
}
