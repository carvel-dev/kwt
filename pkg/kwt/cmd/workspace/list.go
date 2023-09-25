package workspace

import (
	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	ctlwork "github.com/carvel-dev/kwt/pkg/kwt/workspace"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	NamespaceFlags cmdcore.NamespaceFlags
}

func NewListOptions(depsFactory cmdcore.DepsFactory, ui ui.UI) *ListOptions {
	return &ListOptions{depsFactory: depsFactory, ui: ui}
}

func NewListCmd(o *ListOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List all workspaces",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.NamespaceFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *ListOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	ws := ctlwork.NewWorkspaces(o.NamespaceFlags.Name, coreClient)

	wsList, err := ws.List()
	if err != nil {
		return err
	}

	table := uitable.Table{
		Content: "workspaces",

		Header: []uitable.Header{
			uitable.NewHeader("Name"),
			uitable.NewHeader("Alt names"),
			uitable.NewHeader("Ports"),
			uitable.NewHeader("Privileged"),
			uitable.NewHeader("State"),
			uitable.NewHeader("Last used"),
			uitable.NewHeader("Age"),
		},

		SortBy: []uitable.ColumnSort{
			{Column: 0, Asc: true},
		},
	}

	for _, workspace := range wsList {
		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(workspace.Name()),
			cmdcore.NewValueStringsSingleLine(workspace.AltNames()),
			cmdcore.NewValueStringsSingleLine(workspace.Ports()),
			uitable.NewValueBool(workspace.Privileged()),
			uitable.NewValueString(workspace.State()),
			cmdcore.NewValueAge(workspace.LastUsedTime()),
			cmdcore.NewValueAge(workspace.CreationTime()),
		})
	}

	o.ui.PrintTable(table)

	return nil
}
