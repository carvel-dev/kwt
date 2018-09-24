package workspace

import (
	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	ctlwork "github.com/cppforlife/kwt/pkg/kwt/workspace"
	"github.com/spf13/cobra"
)

type EnterOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	WorkspaceFlags WorkspaceFlags
}

func NewEnterOptions(depsFactory cmdcore.DepsFactory, ui ui.UI) *EnterOptions {
	return &EnterOptions{depsFactory: depsFactory, ui: ui}
}

func NewEnterCmd(o *EnterOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "enter",
		Aliases: []string{"e"},
		Short:   "Enter workspace",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.WorkspaceFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *EnterOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	ws := ctlwork.NewWorkspaces(o.WorkspaceFlags.NamespaceFlags.Name, coreClient)

	workspace, err := ws.Find(o.WorkspaceFlags.Name)
	if err != nil {
		return err
	}

	cancelCh := make(chan struct{})

	err = workspace.WaitForStart(cancelCh)
	if err != nil {
		return err
	}

	return workspace.Enter()
}
