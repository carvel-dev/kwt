package e2e

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	// "syscall"
	"testing"
)

type Kwt struct {
	t         *testing.T
	namespace string
	l         Logger
}

type RunOpts struct {
	NoNamespace  bool
	AllowError   bool
	StderrWriter io.Writer
	StdoutWriter io.Writer
	StdinReader  io.Reader
	CancelCh     chan struct{}
	Redact       bool
}

func (k Kwt) Run(args []string) string {
	out, _ := k.RunWithOpts(args, RunOpts{})
	return out
}

func (k Kwt) RunWithOpts(args []string, opts RunOpts) (string, error) {
	k.l.Debugf("Running '%s'...\n", k.cmdDesc(args, opts))

	if !opts.NoNamespace {
		args = append(args, []string{"-n", k.namespace}...)
	}

	cmdName := "kwt"
	cmd := exec.Command(cmdName, args...)

	var stderr, stdout bytes.Buffer

	if opts.StderrWriter != nil {
		cmd.Stderr = opts.StderrWriter
	} else {
		cmd.Stderr = &stderr
	}

	if opts.StdoutWriter != nil {
		cmd.Stdout = opts.StdoutWriter
	} else {
		cmd.Stdout = &stdout
	}

	if opts.CancelCh != nil {
		go func() {
			select {
			case <-opts.CancelCh:
				cmd.Process.Signal(os.Interrupt)
			}
		}()
	}

	err := cmd.Run()
	stdoutStr := stdout.String()

	if err != nil {
		err = fmt.Errorf("Execution error: stdout: '%s' stderr: '%s' error: '%s'", stdoutStr, stderr.String(), err)

		if !opts.AllowError {
			k.t.Fatalf("Failed to successfully execute '%s': %v", k.cmdDesc(args, opts), err)
		}
	}

	return stdoutStr, err
}

func (k Kwt) cmdDesc(args []string, opts RunOpts) string {
	prefix := "kwt"
	if opts.Redact {
		return prefix + " -redacted-"
	}
	return fmt.Sprintf("%s %s", prefix, strings.Join(args, " "))
}
