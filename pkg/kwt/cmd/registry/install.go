package registry

import (
	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	ctlreg "github.com/cppforlife/kwt/pkg/kwt/registry"
	"github.com/spf13/cobra"
)

type InstallOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	Flags Flags
	Debug bool
}

func NewInstallOptions(
	depsFactory cmdcore.DepsFactory,
	ui ui.UI,
) *InstallOptions {
	return &InstallOptions{
		depsFactory: depsFactory,
		ui:          ui,
	}
}

func NewInstallCmd(o *InstallOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"i"},
		Short:   "Install Docker registry into your cluster",
		Example: "kwt registry install",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.Flags.Set(cmd)
	cmd.Flags().BoolVar(&o.Debug, "debug", false, "Set logging level to debug")
	return cmd
}

func (o *InstallOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	logger := cmdcore.NewLoggerWithDebug(o.ui, o.Debug)
	registry := ctlreg.NewRegistry(coreClient, o.Flags.Namespace, logger)

	return registry.Install()
}
