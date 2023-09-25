package workspace

import (
	"time"

	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	ctlwork "github.com/carvel-dev/kwt/pkg/kwt/workspace"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

type InstallOptions struct {
	depsFactory   cmdcore.DepsFactory
	configFactory cmdcore.ConfigFactory
	ui            ui.UI

	WorkspaceFlags WorkspaceFlags
	InstallFlags   InstallFlags
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
	o.InstallFlags.Set("", cmd, flagsFactory)
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

	return InstallOperation{workspace, o.InstallFlags, o.ui, restConfig}.Run()
}

type InstallOperation struct {
	Workspace    ctlwork.Workspace
	InstallFlags InstallFlags
	UI           ui.UI
	RestConfig   *rest.Config
}

type installer struct {
	Enabled     bool
	Title       string
	InstallFunc func(*rest.Config) error
}

func (o InstallOperation) Run() error {
	desktop := ctlwork.WorkspaceDesktop{o.Workspace}

	installers := []installer{
		{Enabled: o.InstallFlags.Desktop, Title: "desktop", InstallFunc: desktop.Install},
		{Enabled: o.InstallFlags.Firefox, Title: "Mozilla Firefox", InstallFunc: desktop.AddFirefox},
		{Enabled: o.InstallFlags.SublimeText, Title: "Sublime Text 3", InstallFunc: desktop.AddSublimeText},
		{Enabled: o.InstallFlags.GoogleChrome, Title: "Google Chrome", InstallFunc: desktop.AddChrome},
		{Enabled: o.InstallFlags.Go1x, Title: "Go 1.x", InstallFunc: desktop.AddGo1x},
		{Enabled: o.InstallFlags.Docker, Title: "Docker", InstallFunc: desktop.AddDocker},
	}

	for _, installer := range installers {
		if installer.Enabled {
			o.UI.PrintLinef("[%s] Installing %s...", time.Now().Format(time.RFC3339), installer.Title)

			err := installer.InstallFunc(o.RestConfig)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
