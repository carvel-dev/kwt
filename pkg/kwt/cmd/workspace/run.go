package workspace

import (
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kwt/pkg/kwt/cmd/core"
	ctlwork "github.com/k14s/kwt/pkg/kwt/workspace"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

type RunOptions struct {
	depsFactory   cmdcore.DepsFactory
	configFactory cmdcore.ConfigFactory
	ui            ui.UI

	WorkspaceFlags WorkspaceFlags
	RunFlags       RunFlags
}

func NewRunOptions(depsFactory cmdcore.DepsFactory, configFactory cmdcore.ConfigFactory, ui ui.UI) *RunOptions {
	return &RunOptions{depsFactory: depsFactory, configFactory: configFactory, ui: ui}
}

func NewRunCmd(o *RunOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Run executable within a workspace",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.WorkspaceFlags.Set(cmd, flagsFactory)
	o.RunFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *RunOptions) Run() error {
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

	return RunOperation{workspace, o.RunFlags, o.ui, restConfig}.Run()
}

type RunOperation struct {
	Workspace  ctlwork.Workspace
	RunFlags   RunFlags
	UI         ui.UI
	RestConfig *rest.Config
}

func (o RunOperation) Run() error {
	err := UploadOperation{o.Workspace, o.RunFlags.SyncFlags.Inputs, o.UI, o.RestConfig}.Run()
	if err != nil {
		return err
	}

	switch {
	case o.RunFlags.Enter:
		return o.Workspace.Enter() // TODO get outputs?

	case o.RunFlags.HasCommand():
		execOpts := ctlwork.ExecuteOpts{
			WorkingDir: o.RunFlags.WorkingDir,

			Command:     []string{"/bin/bash"},
			CommandArgs: []string{"-c", o.RunFlags.Command},
		}

		o.UI.PrintLinef("[%s] Executing command...", time.Now().Format(time.RFC3339))

		err = o.Workspace.Execute(execOpts, o.RestConfig)
		if err != nil {
			return err
		}
	}

	return DownloadOperation{o.Workspace, o.RunFlags.SyncFlags.Outputs, o.UI, o.RestConfig}.Run()
}
