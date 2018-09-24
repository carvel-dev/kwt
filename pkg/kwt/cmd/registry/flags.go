package registry

import (
	"github.com/spf13/cobra"
)

type Flags struct {
	Namespace string
}

func (s *Flags) Set(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.Namespace, "namespace", "kwt-cluster-registry", "Namespace to use to find/install cluster registry")
}
