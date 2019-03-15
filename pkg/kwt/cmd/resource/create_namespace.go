package resource

import (
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/k14s/kwt/pkg/kwt/cmd/core"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateNamespaceOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	NamespaceFlags    cmdcore.NamespaceFlags
	GenerateNameFlags cmdcore.GenerateNameFlags
}

func NewCreateNamespaceOptions(depsFactory cmdcore.DepsFactory, ui ui.UI) *CreateNamespaceOptions {
	return &CreateNamespaceOptions{depsFactory: depsFactory, ui: ui}
}

func NewCreateNamespaceCmd(o *CreateNamespaceOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-namespace",
		Aliases: []string{"create-ns"},
		Short:   "Create namespace",
		Long: `Create namespace.

Use 'kubectl delete ns <name>' to delete namespace.`,
		Example: `
  # Create namespace 'ns1'
  kwt resource create-ns -n ns1`,
		RunE: func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.NamespaceFlags.Set(cmd, flagsFactory)
	o.GenerateNameFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *CreateNamespaceOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	namespace := &corev1.Namespace{
		ObjectMeta: o.GenerateNameFlags.Apply(metav1.ObjectMeta{
			Name: o.NamespaceFlags.Name,
		}),
	}

	createdNamespace, err := coreClient.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		return fmt.Errorf("Creating namespace: %s", err)
	}

	o.printTable(createdNamespace)

	// TODO idempotent?

	return nil
}

func (o *CreateNamespaceOptions) printTable(ns *corev1.Namespace) {
	table := uitable.Table{
		Header: []uitable.Header{
			uitable.NewHeader("Name"),
		},

		Transpose: true,

		Rows: [][]uitable.Value{
			{uitable.NewValueString(ns.Name)},
		},
	}

	o.ui.PrintTable(table)
}
