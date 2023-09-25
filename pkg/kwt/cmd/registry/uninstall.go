package registry

import (
	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	ctlreg "github.com/carvel-dev/kwt/pkg/kwt/registry"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type UninstallOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	Flags Flags
	Debug bool
}

func NewUninstallOptions(
	depsFactory cmdcore.DepsFactory,
	ui ui.UI,
) *UninstallOptions {
	return &UninstallOptions{
		depsFactory: depsFactory,
		ui:          ui,
	}
}

func NewUninstallCmd(o *UninstallOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "uninstall",
		Short:   "Uninstall Docker registry in your cluster",
		Example: "kwt registry uninstall",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.Flags.Set(cmd)
	cmd.Flags().BoolVar(&o.Debug, "debug", false, "Set logging level to debug")
	return cmd
}

func (o *UninstallOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	logger := cmdcore.NewLoggerWithDebug(o.ui, o.Debug)
	registry := ctlreg.NewRegistry(coreClient, o.Flags.Namespace, logger)
	return registry.Uninstall()
}
