package e2e

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	env := BuildEnv(t)
	ctl := Kwt{t, env.Namespace, Logger{}}

	out, _ := ctl.RunWithOpts([]string{"version"}, RunOpts{NoNamespace: true})

	if !strings.Contains(out, "Client Version") {
		t.Fatalf("Expected to find client version")
	}
}
