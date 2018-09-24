package e2e

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLIErrorsForFlagsBeforeExtraArgs(t *testing.T) {
	env := BuildEnv(t)
	ctl := Kwt{t, env.Namespace, Logger{}}

	var stderr bytes.Buffer

	_, err := ctl.RunWithOpts(
		[]string{"workspace", "delete"},
		RunOpts{StderrWriter: &stderr, NoNamespace: true, AllowError: true},
	)

	if err == nil {
		t.Fatalf("Expected to receive error")
	}

	stderrStr := stderr.String()

	// Required flag error is more friendlier than command does not accept extra arg
	if !strings.Contains(stderrStr, `Error: required flag(s) "workspace" not set`) {
		t.Fatalf("Expected to find required flag error, but was '%s'", stderrStr)
	}
}

func TestCLIErrorsForCommandGroups(t *testing.T) {
	env := BuildEnv(t)
	ctl := Kwt{t, env.Namespace, Logger{}}

	// For commands with children commands it's friendlier ux
	// to ignore extra arguments and show available subcommands
	cmdsWithSubcmds := []string{"resource-type", "resource", "net", "workspace", "registry"}

	for _, cmd := range cmdsWithSubcmds {
		var stderr bytes.Buffer

		_, err := ctl.RunWithOpts(
			[]string{cmd, "test-subcmd"},
			RunOpts{StderrWriter: &stderr, NoNamespace: true, AllowError: true},
		)

		if err == nil {
			t.Fatalf("[cmd %s] Expected to receive error", cmd)
		}

		stderrStr := stderr.String()

		if !strings.Contains(stderrStr, "Error: Use one of available subcommands") {
			t.Fatalf("[cmd %s] Expected to find invalid command error in '%s'", cmd, stderrStr)
		}
	}
}
