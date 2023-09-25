package resource

import (
	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	ctlres "github.com/carvel-dev/kwt/pkg/kwt/resources"
	"github.com/spf13/cobra"
)

type ObjectFlags struct {
	Ref ctlres.ObjectRef
}

func (s *ObjectFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	resRef := NewResourceRefFlag(&s.Ref.ResourceRef)
	cmd.Flags().VarP(resRef, "resource-type", "t", "Specified resource type (format: 'group/version/resource')")

	cmd.Flags().StringVarP(&s.Ref.Name, "resource", "r", "", "Specified resource name (format: 'name')")
}
