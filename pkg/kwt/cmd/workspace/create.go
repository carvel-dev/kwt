package workspace

import (
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	ctlwork "github.com/cppforlife/kwt/pkg/kwt/workspace"
	"github.com/spf13/cobra"
)

type CreateOptions struct {
	depsFactory   cmdcore.DepsFactory
	configFactory cmdcore.ConfigFactory
	ui            ui.UI

	WorkspaceFlags WorkspaceFlags
	CreateFlags    CreateFlags
	RunFlags       RunFlags
	DeleteFlags    DeleteFlags
}

func NewCreateOptions(depsFactory cmdcore.DepsFactory, configFactory cmdcore.ConfigFactory, ui ui.UI) *CreateOptions {
	return &CreateOptions{depsFactory: depsFactory, configFactory: configFactory, ui: ui}
}

func NewCreateCmd(o *CreateOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c", "cr"},
		Short:   "Create workspace",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.WorkspaceFlags.SetNonRequired(cmd, flagsFactory)
	o.CreateFlags.Set(cmd, flagsFactory)
	o.RunFlags.Set(cmd, flagsFactory)
	o.DeleteFlags.Set("delete", cmd, flagsFactory)
	return cmd
}

func (o *CreateOptions) Run() error {
	// Generate unique name if no name has been provided
	if len(o.WorkspaceFlags.Name) == 0 {
		o.CreateFlags.CreateOpts.Name = "w"
		o.CreateFlags.CreateOpts.GenerateName = true
	} else {
		o.CreateFlags.CreateOpts.Name = o.WorkspaceFlags.Name
		o.CreateFlags.CreateOpts.GenerateName = o.CreateFlags.GenerateNameFlags.GenerateName
	}

	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	ws := ctlwork.NewWorkspaces(o.WorkspaceFlags.NamespaceFlags.Name, coreClient)

	workspace, err := ws.Create(o.CreateFlags.CreateOpts)
	if err != nil {
		return err
	}

	o.printInfo(workspace)

	defer func() {
		if o.CreateFlags.Remove {
			o.ui.PrintLinef("[%s] Deleting workspace...", time.Now().Format(time.RFC3339))
			workspace.Delete(o.DeleteFlags.Wait) // TODO log error
		}
	}()

	cancelCh := make(chan struct{})

	o.ui.PrintLinef("[%s] Waiting for workspace...", time.Now().Format(time.RFC3339))

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

func (o *CreateOptions) printInfo(workspace ctlwork.Workspace) {
	table := uitable.Table{
		Header: []uitable.Header{
			uitable.NewHeader("Name"),
			uitable.NewHeader("Image"),
			uitable.NewHeader("Ports"),
			uitable.NewHeader("Privileged"),
		},

		Transpose: true,

		Rows: [][]uitable.Value{{
			uitable.NewValueString(workspace.Name()),
			uitable.NewValueString(workspace.Image()),
			uitable.NewValueStrings(workspace.Ports()),
			uitable.NewValueBool(workspace.Privileged()),
		}},
	}

	o.ui.PrintTable(table)
}
