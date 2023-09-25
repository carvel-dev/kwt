package workspace

import (
	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	ctlwork "github.com/carvel-dev/kwt/pkg/kwt/workspace"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type DeleteOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	WorkspaceFlags WorkspaceFlags
	DeleteFlags    DeleteFlags
}

func NewDeleteOptions(depsFactory cmdcore.DepsFactory, ui ui.UI) *DeleteOptions {
	return &DeleteOptions{depsFactory: depsFactory, ui: ui}
}

func NewDeleteCmd(o *DeleteOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"d", "del"},
		Short:   "Delete workspace",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.WorkspaceFlags.Set(cmd, flagsFactory)
	o.DeleteFlags.Set("", cmd, flagsFactory)
	return cmd
}

func (o *DeleteOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	ws := ctlwork.NewWorkspaces(o.WorkspaceFlags.NamespaceFlags.Name, coreClient)

	workspace, err := ws.Find(o.WorkspaceFlags.Name)
	if err != nil {
		return err
	}

	return workspace.Delete(o.DeleteFlags.Wait)
}
