package resourcetype

import (
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/k14s/kwt/pkg/kwt/cmd/core"
	ctlres "github.com/k14s/kwt/pkg/kwt/resources"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI
}

func NewListOptions(depsFactory cmdcore.DepsFactory, ui ui.UI) *ListOptions {
	return &ListOptions{depsFactory, ui}
}

func NewListCmd(o *ListOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all resource types",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	return cmd
}

func (c *ListOptions) Run() error {
	groupHeader := uitable.NewHeader("Group")
	groupHeader.Hidden = true

	versionHeader := uitable.NewHeader("Version")
	versionHeader.Hidden = true

	nameHeader := uitable.NewHeader("Name")
	nameHeader.Hidden = true

	table := uitable.Table{
		Content: "resource types",

		Header: []uitable.Header{
			uitable.NewHeader("Type"),
			groupHeader,
			versionHeader,
			nameHeader,
			uitable.NewHeader("ShortNames"),
			uitable.NewHeader("SingularName"),
			uitable.NewHeader("Namespaced"),
			uitable.NewHeader("Kind"),
			uitable.NewHeader("Verbs"),
			uitable.NewHeader("Categories"),
		},

		SortBy: []uitable.ColumnSort{
			{Column: 1, Asc: true},
			{Column: 2, Asc: true},
			{Column: 3, Asc: true},
		},

		FillFirstColumn: true,
	}

	coreClient, err := c.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	pairs, err := ctlres.NewResourceTypes(coreClient).All()
	if err != nil {
		return err
	}

	for _, pair := range pairs {
		group := pair.GroupVersionResource.Group
		if len(pair.APIResource.Group) > 0 {
			group = pair.APIResource.Group
		}

		version := pair.GroupVersionResource.Version
		if len(pair.APIResource.Version) > 0 {
			version = pair.APIResource.Version
		}

		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(fmt.Sprintf("%s/%s/%s", group, version, pair.APIResource.Name)),
			uitable.NewValueString(group),
			uitable.NewValueString(version),
			uitable.NewValueString(pair.APIResource.Name),
			cmdcore.NewValueStringsSingleLine(pair.APIResource.ShortNames),
			uitable.NewValueString(pair.APIResource.SingularName),
			uitable.NewValueBool(pair.APIResource.Namespaced),
			uitable.NewValueString(pair.APIResource.Kind),
			cmdcore.NewValueStringsSingleLine(pair.APIResource.Verbs),
			cmdcore.NewValueStringsSingleLine(pair.APIResource.Categories),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
