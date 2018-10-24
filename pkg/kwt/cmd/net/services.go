package net

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	ctlkubedns "github.com/cppforlife/kwt/pkg/kwt/kubedns"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServicesOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	NamespaceFlags cmdcore.NamespaceFlags
}

func NewServicesOptions(depsFactory cmdcore.DepsFactory, ui ui.UI) *ServicesOptions {
	return &ServicesOptions{depsFactory: depsFactory, ui: ui}
}

func NewServicesCmd(o *ServicesOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "services",
		Aliases: []string{"svc", "svcs", "service"},
		Short:   "List all services",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.NamespaceFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *ServicesOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	svcList, err := coreClient.CoreV1().Services(o.NamespaceFlags.Name).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	table := uitable.Table{
		Title: fmt.Sprintf("Services in namespace '%s'", o.NamespaceFlags.Name),

		Content: "services",

		Header: []uitable.Header{
			uitable.NewHeader("Name"),
			uitable.NewHeader("Internal DNS"),
			uitable.NewHeader("Cluster IP"),
			uitable.NewHeader("Ports"),
		},

		SortBy: []uitable.ColumnSort{
			{Column: 0, Asc: true},
		},
	}

	resolver := ctlkubedns.NewKubeDNSIPResolver(ctlkubedns.DefaultClusterDomain, coreClient)

	for _, svc := range svcList.Items {
		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(svc.Name),
			uitable.NewValueString(resolver.ServiceInternalDNSAddress(svc)),
			uitable.NewValueString(svc.Spec.ClusterIP),
			cmdcore.NewValueStringsSingleLine(o.svcPorts(svc)),
		})
	}

	o.ui.PrintTable(table)

	return nil
}

func (o *ServicesOptions) svcPorts(svc corev1.Service) []string {
	var result []string
	for _, port := range svc.Spec.Ports {
		result = append(result, strconv.Itoa(int(port.Port))+"/"+strings.ToLower(string(port.Protocol)))
	}
	return result
}
