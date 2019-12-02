package e2e

import (
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

type KwtNet struct {
	cancelCh        chan struct{}
	doneCh          chan struct{}
	collectedOutput *LogsWriter

	kwt    Kwt
	t      *testing.T
	logger Logger
}

func NewKwtNet(kwt Kwt, t *testing.T, logger Logger) *KwtNet {
	return &KwtNet{
		cancelCh:        make(chan struct{}),
		doneCh:          make(chan struct{}),
		collectedOutput: &LogsWriter{},

		kwt:    kwt,
		t:      t,
		logger: logger,
	}
}

func (k *KwtNet) CollectedOutput() string {
	return k.collectedOutput.Current()
}

func (k *KwtNet) Start(args []string) {
	k.cleanUp()
	k.StartWithoutCleanup(args)
}

func (k *KwtNet) Listen(args []string) {
	k.cleanUp()

	k.logger.Section("Starting net listen in background", func() {
		go func() {
			k.kwt.RunWithOpts(append([]string{"net", "listen", "--tty"}, args...), RunOpts{StdoutWriter: k.collectedOutput, CancelCh: k.cancelCh})
			k.doneCh <- struct{}{}
		}()
	})

	k.waitForReady()
}

func (k *KwtNet) End() {
	k.EndWithoutCleanup()
	k.cleanUp()
}

func (k *KwtNet) StartWithoutCleanup(args []string) {
	k.logger.Section("Starting net start in background", func() {
		go func() {
			k.kwt.RunWithOpts(append([]string{"net", "start", "--tty"}, args...), RunOpts{StdoutWriter: k.collectedOutput, CancelCh: k.cancelCh, NoNamespace: true})
			k.doneCh <- struct{}{}
		}()
	})

	k.waitForReady()
}

func (k *KwtNet) EndWithoutCleanup() {
	k.logger.Section("Terminating net command tailing", func() {
		k.cancelCh <- struct{}{}
		<-k.doneCh
	})
}

func (k *KwtNet) waitForReady() {
	k.logger.Section("Wait for forwarding to be ready", func() {
		timeoutCh := time.After(2 * time.Minute)
		const expectedOutput = "Ready!"

		for {
			currOutput := k.collectedOutput.Current()

			if strings.Contains(currOutput, expectedOutput) {
				break
			}

			select {
			case <-timeoutCh:
				k.t.Fatalf("Timed out waiting for '%s' to be seen in output '%s'", expectedOutput, currOutput)
			default:
				// continue with waiting
			}

			time.Sleep(1 * time.Second)
		}
	})
}

func (k *KwtNet) cleanUp() {
	k.logger.Section("Clean up net access endpoint", func() {
		k.kwt.RunWithOpts([]string{"net", "clean-up"}, RunOpts{AllowError: true, NoNamespace: true})
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
