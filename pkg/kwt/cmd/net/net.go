package net

import (
	"github.com/spf13/cobra"
)

func NewNetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "net",
		Aliases: []string{"n"},
		Short:   "Network",
	}
	return cmd
}
