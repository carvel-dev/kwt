package workspace

import (
	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kwt/pkg/kwt/cmd/core"
	ctlwork "github.com/k14s/kwt/pkg/kwt/workspace"
	"github.com/spf13/cobra"
)

type AddAltNameOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	WorkspaceFlags WorkspaceFlags
	AltName        string
}

func NewAddAltNameOptions(depsFactory cmdcore.DepsFactory, ui ui.UI) *AddAltNameOptions {
	return &AddAltNameOptions{depsFactory: depsFactory, ui: ui}
}

func NewAddAltNameCmd(o *AddAltNameOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-alt-name",
		Aliases: []string{"aan"},
		Short:   "Add alternative name to workspace",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.WorkspaceFlags.Set(cmd, flagsFactory)
	cmd.Flags().StringVarP(&o.AltName, "name", "a", "", "Specified alternative name")
	return cmd
}

func (o *AddAltNameOptions) Run() error {
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

	return workspace.AddAltName(o.AltName)
}
