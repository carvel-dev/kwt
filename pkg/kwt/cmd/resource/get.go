package resource

import (
	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kwt/pkg/kwt/cmd/core"
	ctlres "github.com/k14s/kwt/pkg/kwt/resources"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type GetOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	NamespaceFlags cmdcore.NamespaceFlags
	ObjectFlags    ObjectFlags
}

func NewGetOptions(depsFactory cmdcore.DepsFactory, ui ui.UI) *GetOptions {
	return &GetOptions{depsFactory: depsFactory, ui: ui}
}

func NewGetCmd(o *GetOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get resource",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.NamespaceFlags.Set(cmd, flagsFactory)
	o.ObjectFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *GetOptions) Run() error {
	dynamicClient, err := o.depsFactory.DynamicClient()
	if err != nil {
		return err
	}

	item, err := ctlres.NewResources(nil, dynamicClient).Get(o.ObjectFlags.Ref, o.NamespaceFlags.Name)
	if err != nil {
		return err
	}

	bytes, err := yaml.Marshal(item.Item.UnstructuredContent())
	if err != nil {
		return err
	}

	o.ui.PrintBlock(bytes)

	return nil
}
