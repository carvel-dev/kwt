package registry

import (
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	ctlreg "github.com/cppforlife/kwt/pkg/kwt/registry"
	"github.com/spf13/cobra"
)

type InfoOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	Flags Flags
}

func NewInfoOptions(
	depsFactory cmdcore.DepsFactory,
	ui ui.UI,
) *InfoOptions {
	return &InfoOptions{
		depsFactory: depsFactory,
		ui:          ui,
	}
}

func NewInfoCmd(o *InfoOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info",
		Short:   "Show info about Docker registry in your cluster",
		Example: "kwt registry info",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.Flags.Set(cmd)
	return cmd
}

func (o *InfoOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	logger := cmdcore.NewLogger(o.ui)
	registry := ctlreg.NewRegistry(coreClient, o.Flags.Namespace, logger)

	info, err := registry.Info()
	if err != nil {
		return err
	}

	table := uitable.Table{
		Header: []uitable.Header{
			uitable.NewHeader("Address"),
			uitable.NewHeader("Username"),
			uitable.NewHeader("Password"),
			uitable.NewHeader("CA"),
		},

		Transpose: true,
	}

	table.Rows = append(table.Rows, []uitable.Value{
		uitable.NewValueString(info.Address),
		uitable.NewValueString(info.Username),
		uitable.NewValueString(info.Password),
		uitable.NewValueString(info.CA),
	})

	o.ui.PrintTable(table)

	return nil
}
