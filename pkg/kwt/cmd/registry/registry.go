package registry

import (
	"github.com/spf13/cobra"
)

func NewRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Registry",
	}
	return cmd
}
