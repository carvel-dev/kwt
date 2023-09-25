package resource

import (
	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	ctlres "github.com/carvel-dev/kwt/pkg/kwt/resources"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type DeleteOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	NamespaceFlags cmdcore.NamespaceFlags
	ObjectFlags    ObjectFlags
}

func NewDeleteOptions(depsFactory cmdcore.DepsFactory, ui ui.UI) *DeleteOptions {
	return &DeleteOptions{depsFactory: depsFactory, ui: ui}
}

func NewDeleteCmd(o *DeleteOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resource",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.NamespaceFlags.Set(cmd, flagsFactory)
	o.ObjectFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *DeleteOptions) Run() error {
	dynamicClient, err := o.depsFactory.DynamicClient()
	if err != nil {
		return err
	}

	return ctlres.NewResources(nil, dynamicClient).Delete(o.ObjectFlags.Ref, o.NamespaceFlags.Name)
}
