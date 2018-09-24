package e2e

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNet(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	ctl := Kwt{t, env.Namespace, Logger{}}

	cleanUp := func() {
		logger.Section("Clean up net access endpoint", func() {
			ctl.RunWithOpts([]string{"net", "clean-up"}, RunOpts{AllowError: true, NoNamespace: true})
		})
	}

	cleanUp()
	defer cleanUp()

	cancelCh := make(chan struct{})
	doneCh := make(chan struct{})
	collectedOutput := &LogsWriter{}

	logger.Section("Starting net start in background", func() {
		go func() {
			ctl.RunWithOpts([]string{"net", "start", "--tty"}, RunOpts{StdoutWriter: collectedOutput, CancelCh: cancelCh, NoNamespace: true})
			doneCh <- struct{}{}
		}()
	})

	logger.Section("Wait for forwarding to be ready", func() {
		timeoutCh := time.After(2 * time.Minute)
		const expectedOutput = "Ready!"

		for {
			if strings.Contains(collectedOutput.Current(), expectedOutput) {
				break
			}

			select {
			case <-timeoutCh:
				t.Fatalf("Timed out waiting for '%s' to be seen in output", expectedOutput)
			default:
				// continue with waiting
			}

			time.Sleep(1 * time.Second)
		}
	})

	logger.Section("Test network accessibility to the service", func() {
		const expectedOutput = "Hello Kubernetes!"

		res, err := http.Get(fmt.Sprintf("http://%s", env.ServiceClusterIP))
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

	logger.Section("Terminating net start tailing", func() {
		cancelCh <- struct{}{}
		<-doneCh
	})
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
