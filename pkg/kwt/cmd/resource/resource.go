package resource

import (
	"github.com/spf13/cobra"
)

func NewResourceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource",
		Aliases: []string{"r"},
		Short:   "Resource",
	}
	return cmd
}
