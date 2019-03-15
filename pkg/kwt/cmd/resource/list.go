package resource

import (
	"fmt"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/k14s/kwt/pkg/kwt/cmd/core"
	ctlres "github.com/k14s/kwt/pkg/kwt/resources"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ListOptions struct {
	depsFactory cmdcore.DepsFactory
	ui          ui.UI

	NamespaceFlags      cmdcore.NamespaceFlags
	ResourceFilterFlags ResourceFilterFlags
}

func NewListOptions(depsFactory cmdcore.DepsFactory, ui ui.UI) *ListOptions {
	return &ListOptions{depsFactory: depsFactory, ui: ui}
}

func NewListCmd(o *ListOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all resources",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.NamespaceFlags.Set(cmd, flagsFactory)
	o.ResourceFilterFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *ListOptions) Run() error {
	statusHeader := uitable.NewHeader("Status")
	statusHeader.Hidden = true

	labelsHeader := uitable.NewHeader("Labels")
	labelsHeader.Hidden = true

	annsHeader := uitable.NewHeader("Annotations")
	annsHeader.Hidden = true

	table := uitable.Table{
		Content: "resources",

		Header: []uitable.Header{
			uitable.NewHeader("Type"),
			uitable.NewHeader("Namespace"),
			uitable.NewHeader("Name"),
			uitable.NewHeader("Conditions"),
			statusHeader,
			labelsHeader,
			annsHeader,
		},

		SortBy: []uitable.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 1, Asc: true},
			{Column: 2, Asc: true},
		},

		FillFirstColumn: true,
	}

	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	resTypes, err := ctlres.NewResourceTypes(coreClient).All()
	if err != nil {
		return err
	}

	allOpts := ctlres.ResourcesAllOpts{}

	if len(o.NamespaceFlags.Name) > 0 {
		allOpts = ctlres.ResourcesAllOpts{
			IncludeNonNamespaced: false,
			IncludeNamespaces:    []string{o.NamespaceFlags.Name},
		}
	} else {
		allOpts = ctlres.ResourcesAllOpts{
			IncludeNonNamespaced: true,
			IncludeAllNamespaces: true,
		}
	}

	resTypes = ctlres.Listable(resTypes)

	// TODO if o.ResourceFilterFlags.ObjectFlags.ResourceRef != nil {
	// 	resTypes = ctlres.Matching(resTypes, *opts.Args.ResourceRef)
	// }

	// Exclude events since they are pretty noisy
	resTypes = ctlres.NonMatching(resTypes, ctlres.ResourceRef{
		schema.GroupVersionResource{Version: "v1", Resource: "events"},
	})

	dynamicClient, err := o.depsFactory.DynamicClient()
	if err != nil {
		return err
	}

	resources, err := ctlres.NewResources(coreClient, dynamicClient).All(resTypes, allOpts)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		content := resource.Item.UnstructuredContent()
		name := ""

		var labelsVal, annsVal uitable.Value

		if meta, ok := content["metadata"]; ok {
			if typedMeta, ok := meta.(map[string]interface{}); ok {
				name = typedMeta["name"].(string)

				if len(o.ResourceFilterFlags.ObjectFlags.Ref.Name) > 0 {
					if name != o.ResourceFilterFlags.ObjectFlags.Ref.Name {
						continue
					}
				}

				labelsVal = NewLabelsTableValue(typedMeta)
				annsVal = NewAnnotationsTableValue(typedMeta)

				if len(o.ResourceFilterFlags.NameIncludes) > 0 {
					if !strings.Contains(name, o.ResourceFilterFlags.NameIncludes) {
						continue
					}
				}

				if len(o.ResourceFilterFlags.LabelIncludes) > 0 {
					var matches bool
					if typedLabels, ok := typedMeta["labels"].(map[string]interface{}); ok {
						for k, _ := range typedLabels {
							if strings.Contains(k, o.ResourceFilterFlags.LabelIncludes) {
								matches = true
								break
							}
						}
					}
					if !matches {
						continue
					}
				}

				if len(o.ResourceFilterFlags.AnnotationIncludes) > 0 {
					var matches bool
					if typedAnnotations, ok := typedMeta["annotations"].(map[string]interface{}); ok {
						for k, _ := range typedAnnotations {
							if strings.Contains(k, o.ResourceFilterFlags.AnnotationIncludes) {
								matches = true
								break
							}
						}
					}
					if !matches {
						continue
					}
				}
			}
		}

		var conditionsVal, statusVal uitable.Value

		if status, ok := content["status"].(map[string]interface{}); ok {
			plainConditionsVal := NewShortConditionsTableValue(status)
			conditionsVal = uitable.ValueFmt{V: plainConditionsVal, Error: plainConditionsVal.NeedsAttention()}
			statusVal = NewStatusTableValue(status)
		} else {
			conditionsVal = uitable.ValueFmt{V: uitable.NewValueInterface(nil), Error: false}
			statusVal = uitable.NewValueInterface(nil)
		}

		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(fmt.Sprintf("%s/%s/%s", resource.GroupVersionResource.Group, resource.GroupVersionResource.Version, resource.GroupVersionResource.Resource)),
			uitable.NewValueString(resource.Namespace),
			uitable.NewValueString(name),
			conditionsVal,
			statusVal,
			labelsVal,
			annsVal,
		})
	}

	o.ui.PrintTable(table)

	return nil
}
