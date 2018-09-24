package workspace

import (
	"github.com/spf13/cobra"
)

func NewWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"w"},
		Short:   "Workspace",
	}
	return cmd
}
