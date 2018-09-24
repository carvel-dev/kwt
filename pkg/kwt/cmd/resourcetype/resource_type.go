package resourcetype

import (
	"github.com/spf13/cobra"
)

func NewResourceTypeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource-type",
		Aliases: []string{"rt"},
		Short:   "Resource type",
	}
	return cmd
}
