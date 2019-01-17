package workspace

import (
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	ctlwork "github.com/cppforlife/kwt/pkg/kwt/workspace"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

type InstallOptions struct {
	depsFactory   cmdcore.DepsFactory
	configFactory cmdcore.ConfigFactory
	ui            ui.UI

	WorkspaceFlags WorkspaceFlags

	Desktop      bool
	Firefox      bool
	SublimeText  bool
	GoogleChrome bool
	Go1x         bool
}

type installer struct {
	Enabled     bool
	Title       string
	InstallFunc func(*rest.Config) error
}

func NewInstallOptions(depsFactory cmdcore.DepsFactory, configFactory cmdcore.ConfigFactory, ui ui.UI) *InstallOptions {
	return &InstallOptions{depsFactory: depsFactory, configFactory: configFactory, ui: ui}
}

func NewInstallCmd(o *InstallOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"i"},
		Short:   "Install predefined application into workspace",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.WorkspaceFlags.Set(cmd, flagsFactory)

	cmd.Flags().BoolVar(&o.Desktop, "desktop", false, "Configure X11 and VNC access")
	cmd.Flags().BoolVar(&o.Firefox, "firefox", false, "Install Firefox")
	cmd.Flags().BoolVar(&o.SublimeText, "sublime", false, "Install Sublime Text")
	cmd.Flags().BoolVar(&o.GoogleChrome, "chrome", false, "Install Google Chrome")
	cmd.Flags().BoolVar(&o.Go1x, "go1x", false, "Install Go 1.x")

	return cmd
}

func (o *InstallOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	ws := ctlwork.NewWorkspaces(o.WorkspaceFlags.NamespaceFlags.Name, coreClient)

	workspace, err := ws.Find(o.WorkspaceFlags.Name)
	if err != nil {
		return err
	}

	_ = workspace.MarkUse()

	cancelCh := make(chan struct{})

	err = workspace.WaitForStart(cancelCh)
	if err != nil {
		return err
	}

	restConfig, err := o.configFactory.RESTConfig()
	if err != nil {
		return err
	}

	desktop := ctlwork.WorkspaceDesktop{workspace}

	installers := []installer{
		{Enabled: o.Desktop, Title: "desktop", InstallFunc: desktop.Install},
		{Enabled: o.Firefox, Title: "Mozilla Firefox", InstallFunc: desktop.AddFirefox},
		{Enabled: o.SublimeText, Title: "Sublime Text 3", InstallFunc: desktop.AddSublimeText},
		{Enabled: o.GoogleChrome, Title: "Google Chrome", InstallFunc: desktop.AddChrome},
		{Enabled: o.Go1x, Title: "Go 1.x", InstallFunc: desktop.AddGo1x},
	}

	for _, installer := range installers {
		if installer.Enabled {
			o.ui.PrintLinef("[%s] Installing %s...", time.Now().Format(time.RFC3339), installer.Title)

			err := installer.InstallFunc(restConfig)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
