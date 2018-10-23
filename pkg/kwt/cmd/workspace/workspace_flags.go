package workspace

import (
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	"github.com/spf13/cobra"
)

type WorkspaceFlags struct {
	NamespaceFlags cmdcore.NamespaceFlags
	Name           string
}

func (s *WorkspaceFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	s.SetNonRequired(cmd, flagsFactory)
	cmd.MarkFlagRequired("workspace")
}

func (s *WorkspaceFlags) SetNonRequired(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	s.NamespaceFlags.Set(cmd, flagsFactory)

	cmd.Flags().StringVarP(&s.Name, "workspace", "w", "", "Specified workspace")
}
