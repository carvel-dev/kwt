package net

import (
	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kwt/pkg/kwt/cmd/core"
	ctlnet "github.com/k14s/kwt/pkg/kwt/net"
	"github.com/spf13/cobra"
)

type CleanUpOptions struct {
	depsFactory   cmdcore.DepsFactory
	configFactory cmdcore.ConfigFactory
	ui            ui.UI

	NamespaceFlags NamespaceFlags
	LoggingFlags   LoggingFlags
}

func NewCleanUpOptions(
	depsFactory cmdcore.DepsFactory,
	configFactory cmdcore.ConfigFactory,
	ui ui.UI,
) *CleanUpOptions {
	return &CleanUpOptions{depsFactory: depsFactory, configFactory: configFactory, ui: ui}
}

func NewCleanUpCmd(o *CleanUpOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clean-up",
		Aliases: []string{"cleanup"},
		Short:   "Clean up network access",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.NamespaceFlags.Set(cmd)
	o.LoggingFlags.Set(cmd)
	return cmd
}

func (o *CleanUpOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	restConfig, err := o.configFactory.RESTConfig()
	if err != nil {
		return err
	}

	logger := cmdcore.NewLoggerWithDebug(o.ui, o.LoggingFlags.Debug)

	return ctlnet.NewKubeEntryPoint(coreClient, restConfig, o.NamespaceFlags.Name, logger).Delete()
}
