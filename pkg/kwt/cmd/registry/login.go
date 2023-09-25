package registry

import (
	"fmt"
	"os/exec"
	"strings"

	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	ctlreg "github.com/carvel-dev/kwt/pkg/kwt/registry"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type LogInOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	Flags Flags
}

func NewLogInOptions(
	depsFactory cmdcore.DepsFactory,
	ui ui.UI,
) *LogInOptions {
	return &LogInOptions{
		depsFactory: depsFactory,
		ui:          ui,
	}
}

func NewLogInCmd(o *LogInOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log-in",
		Aliases: []string{"l", "login"},
		Short:   "Log in Docker daemon to use cluster registry",
		Example: "kwt registry l",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.Flags.Set(cmd)
	return cmd
}

func (o *LogInOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	logger := cmdcore.NewLogger(o.ui)
	registry := ctlreg.NewRegistry(coreClient, o.Flags.Namespace, logger)

	info, err := registry.Info()
	if err != nil {
		return err
	}

	cmd := exec.Command("docker", "login", "https://"+info.Address, "-u", info.Username, "--password-stdin")
	cmd.Stdin = strings.NewReader(info.Password)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running docker login: %s output: %s", err, output)
	}

	o.ui.PrintLinef("Successfully logged in Docker daemon to registry '%s'", info.Address)

	return nil
}
