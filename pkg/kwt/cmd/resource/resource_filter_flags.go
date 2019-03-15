package resource

import (
	cmdcore "github.com/k14s/kwt/pkg/kwt/cmd/core"
	"github.com/spf13/cobra"
)

type ResourceFilterFlags struct {
	ObjectFlags

	NameIncludes       string
	LabelIncludes      string
	AnnotationIncludes string
	// TODO -l https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
}

func (s *ResourceFilterFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	s.ObjectFlags.Set(cmd, flagsFactory)

	cmd.Flags().StringVar(&s.NameIncludes, "name-includes", "", "Filter by content in resource name")
	cmd.Flags().StringVar(&s.LabelIncludes, "label-includes", "", "Filter by content in resource label keys")
	cmd.Flags().StringVar(&s.AnnotationIncludes, "add-includes", "", "Filter by content in resource annotation keys")
}
