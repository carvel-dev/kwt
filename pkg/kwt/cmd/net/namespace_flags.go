package net

import (
	"github.com/spf13/cobra"
)

type NamespaceFlags struct {
	Name string
}

func (s *NamespaceFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.Name, "namespace", "n", "default", "Namespace to use to manage networking pod")
}
