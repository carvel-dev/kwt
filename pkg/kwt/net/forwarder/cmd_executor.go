package forwarder

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type CmdExecutor interface {
	CombinedOutput(cmdName string, args []string, stdin io.Reader) ([]byte, error)
}

type OsCmdExecutor struct {
	logger Logger
	logTag string
}

var _ CmdExecutor = OsCmdExecutor{}

func NewOsCmdExecutor(logger Logger) OsCmdExecutor {
	return OsCmdExecutor{logger, "OsCmdExecutor"}
}

func (e OsCmdExecutor) CombinedOutput(cmdName string, args []string, stdin io.Reader) ([]byte, error) {
	cmdDesc := cmdName + " " + strings.Join(args, " ")
	e.logger.Debug(e.logTag, "Running '%s'", cmdDesc)

	cmd := exec.Command("iptables", append([]string{"-w"}, args...)...)
	cmd.Stdin = stdin

	out, err := cmd.CombinedOutput()
	if err != nil {
		e.logger.Debug(e.logTag, "Failed, error: %s, output: %s", err, string(out))
		return out, fmt.Errorf("Running '%s': %s", cmdDesc, err)
	} else {
		e.logger.Debug(e.logTag, "Succeeded, output: %s", string(out))
	}

	return out, nil
}
