package net

import (
	"fmt"
	"strconv"
	"strings"

	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	ctlkubedns "github.com/carvel-dev/kwt/pkg/kwt/kubedns"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodsOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	NamespaceFlags cmdcore.NamespaceFlags
}

func NewPodsOptions(depsFactory cmdcore.DepsFactory, ui ui.UI) *PodsOptions {
	return &PodsOptions{depsFactory: depsFactory, ui: ui}
}

func NewPodsCmd(o *PodsOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pods",
		Aliases: []string{"pod"},
		Short:   "List all pods",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.NamespaceFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *PodsOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	podList, err := coreClient.CoreV1().Pods(o.NamespaceFlags.Name).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	table := uitable.Table{
		Title: fmt.Sprintf("Pods in namespace '%s'", o.NamespaceFlags.Name),

		Content: "pods",

		Header: []uitable.Header{
			uitable.NewHeader("Name"),
			uitable.NewHeader("Internal DNS"),
			uitable.NewHeader("IP"),
			uitable.NewHeader("Ports"),
		},

		SortBy: []uitable.ColumnSort{
			{Column: 0, Asc: true},
		},
	}

	resolver := ctlkubedns.NewKubeDNSIPResolver(ctlkubedns.DefaultClusterDomain, coreClient)

	for _, pod := range podList.Items {
		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(pod.Name),
			uitable.NewValueString(resolver.PodInternalDNSAddress(pod)),
			uitable.NewValueString(pod.Status.PodIP),
			cmdcore.NewValueStringsSingleLine(o.podPorts(pod)),
		})
	}

	o.ui.PrintTable(table)

	return nil
}

func (o *PodsOptions) podPorts(pod corev1.Pod) []string {
	var result []string
	for _, cont := range pod.Spec.Containers {
		for _, port := range cont.Ports {
			result = append(result, strconv.Itoa(int(port.ContainerPort))+"/"+strings.ToLower(string(port.Protocol)))
		}
	}
	return result
}
