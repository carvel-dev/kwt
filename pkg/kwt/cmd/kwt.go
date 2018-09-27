package cmd

import (
	"io"

	"github.com/cppforlife/cobrautil"
	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	cmdnet "github.com/cppforlife/kwt/pkg/kwt/cmd/net"
	cmdreg "github.com/cppforlife/kwt/pkg/kwt/cmd/registry"
	cmdres "github.com/cppforlife/kwt/pkg/kwt/cmd/resource"
	cmdrestype "github.com/cppforlife/kwt/pkg/kwt/cmd/resourcetype"
	cmdwork "github.com/cppforlife/kwt/pkg/kwt/cmd/workspace"
	"github.com/spf13/cobra"
)

type KwtOptions struct {
	ui            *ui.ConfUI
	configFactory cmdcore.ConfigFactory
	depsFactory   cmdcore.DepsFactory

	UIFlags         UIFlags
	KubeconfigFlags cmdcore.KubeconfigFlags
}

func NewKwtOptions(ui *ui.ConfUI, configFactory cmdcore.ConfigFactory, depsFactory cmdcore.DepsFactory) *KwtOptions {
	return &KwtOptions{ui: ui, configFactory: configFactory, depsFactory: depsFactory}
}

func NewDefaultKwtCmd(ui *ui.ConfUI) *cobra.Command {
	configFactory := cmdcore.NewConfigFactoryImpl()
	depsFactory := cmdcore.NewDepsFactoryImpl(configFactory)
	options := NewKwtOptions(ui, configFactory, depsFactory)
	flagsFactory := cmdcore.NewFlagsFactory(configFactory, depsFactory)
	return NewKwtCmd(options, flagsFactory)
}

func NewKwtCmd(o *KwtOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kwt",
		Short: "kwt helps develop with your Kubernetes cluster",
		Long:  "kwt helps develop with your Kubernetes cluster.",

		RunE: cobrautil.ShowHelp,

		// Affects children as well
		SilenceErrors: true,
		SilenceUsage:  true,

		// Disable docs header
		DisableAutoGenTag: true,

		// TODO bash completion
	}

	cmd.SetOutput(uiBlockWriter{o.ui}) // setting output for cmd.Help()

	o.UIFlags.Set(cmd, flagsFactory)
	o.KubeconfigFlags.Set(cmd, flagsFactory)

	o.configFactory.ConfigurePathResolver(o.KubeconfigFlags.Path.Value)
	o.configFactory.ConfigureContextResolver(o.KubeconfigFlags.Context.Value)

	cancelSignals := cmdcore.CancelSignals{}

	cmd.AddCommand(NewVersionCmd(NewVersionOptions(o.ui), flagsFactory))

	resourceTypeCmd := cmdrestype.NewResourceTypeCmd()
	resourceTypeCmd.Hidden = true
	resourceTypeCmd.AddCommand(cmdrestype.NewListCmd(cmdrestype.NewListOptions(o.depsFactory, o.ui), flagsFactory))
	cmd.AddCommand(resourceTypeCmd)

	resourceCmd := cmdres.NewResourceCmd()
	resourceCmd.Hidden = true
	resourceCmd.AddCommand(cmdres.NewListCmd(cmdres.NewListOptions(o.depsFactory, o.ui), flagsFactory))
	resourceCmd.AddCommand(cmdres.NewGetCmd(cmdres.NewGetOptions(o.depsFactory, o.ui), flagsFactory))
	resourceCmd.AddCommand(cmdres.NewCreateNamespaceCmd(cmdres.NewCreateNamespaceOptions(o.depsFactory, o.ui), flagsFactory))
	resourceCmd.AddCommand(cmdres.NewDeleteCmd(cmdres.NewDeleteOptions(o.depsFactory, o.ui), flagsFactory))
	cmd.AddCommand(resourceCmd)

	workspaceCmd := cmdwork.NewWorkspaceCmd()
	workspaceCmd.AddCommand(cmdwork.NewListCmd(cmdwork.NewListOptions(o.depsFactory, o.ui), flagsFactory))
	workspaceCmd.AddCommand(cmdwork.NewCreateCmd(cmdwork.NewCreateOptions(o.depsFactory, o.configFactory, o.ui), flagsFactory))
	workspaceCmd.AddCommand(cmdwork.NewDeleteCmd(cmdwork.NewDeleteOptions(o.depsFactory, o.ui), flagsFactory))
	workspaceCmd.AddCommand(cmdwork.NewEnterCmd(cmdwork.NewEnterOptions(o.depsFactory, o.ui), flagsFactory))
	workspaceCmd.AddCommand(cmdwork.NewSyncCmd(cmdwork.NewSyncOptions(o.depsFactory, o.configFactory, o.ui), flagsFactory))
	workspaceCmd.AddCommand(cmdwork.NewRunCmd(cmdwork.NewRunOptions(o.depsFactory, o.configFactory, o.ui), flagsFactory))
	workspaceCmd.AddCommand(cmdwork.NewInstallCmd(cmdwork.NewInstallOptions(o.depsFactory, o.configFactory, o.ui), flagsFactory))
	workspaceCmd.AddCommand(cmdwork.NewAddAltNameCmd(cmdwork.NewAddAltNameOptions(o.depsFactory, o.ui), flagsFactory))
	cmd.AddCommand(workspaceCmd)

	netCmd := cmdnet.NewNetCmd()
	netCmd.AddCommand(cmdnet.NewStartCmd(cmdnet.NewStartOptions(o.depsFactory, o.configFactory, o.ui, cancelSignals), flagsFactory))
	netCmd.AddCommand(cmdnet.NewCleanUpCmd(cmdnet.NewCleanUpOptions(o.depsFactory, o.configFactory, o.ui), flagsFactory))
	netCmd.AddCommand(cmdnet.NewForwardCmd(cmdnet.NewForwardOptions(o.depsFactory, o.ui, cancelSignals), flagsFactory))
	netCmd.AddCommand(cmdnet.NewServicesCmd(cmdnet.NewServicesOptions(o.depsFactory, o.ui), flagsFactory))
	netCmd.AddCommand(cmdnet.NewPodsCmd(cmdnet.NewPodsOptions(o.depsFactory, o.ui), flagsFactory))
	netCmd.AddCommand(cmdnet.NewStartDNSCmd(cmdnet.NewStartDNSOptions(o.depsFactory, o.ui, cancelSignals), flagsFactory))
	cmd.AddCommand(netCmd)

	regCmd := cmdreg.NewRegistryCmd()
	regCmd.Hidden = true
	regCmd.AddCommand(cmdreg.NewInstallCmd(cmdreg.NewInstallOptions(o.depsFactory, o.ui), flagsFactory))
	regCmd.AddCommand(cmdreg.NewInfoCmd(cmdreg.NewInfoOptions(o.depsFactory, o.ui), flagsFactory))
	regCmd.AddCommand(cmdreg.NewLogInCmd(cmdreg.NewLogInOptions(o.depsFactory, o.ui), flagsFactory))
	regCmd.AddCommand(cmdreg.NewUninstallCmd(cmdreg.NewUninstallOptions(o.depsFactory, o.ui), flagsFactory))
	cmd.AddCommand(regCmd)

	// Last one runs first
	cobrautil.VisitCommands(cmd, cobrautil.ReconfigureCmdWithSubcmd)
	cobrautil.VisitCommands(cmd, cobrautil.ReconfigureLeafCmd)

	cobrautil.VisitCommands(cmd, cobrautil.WrapRunEForCmd(func(*cobra.Command, []string) error {
		o.UIFlags.ConfigureUI(o.ui)
		return nil
	}))

	cobrautil.VisitCommands(cmd, cobrautil.WrapRunEForCmd(cobrautil.ResolveFlagsForCmd))

	return cmd
}

type uiBlockWriter struct {
	ui ui.UI
}

var _ io.Writer = uiBlockWriter{}

func (w uiBlockWriter) Write(p []byte) (n int, err error) {
	w.ui.PrintBlock(p)
	return len(p), nil
}
