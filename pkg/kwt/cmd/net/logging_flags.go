package net

import (
	"github.com/spf13/cobra"
)

type LoggingFlags struct {
	Debug bool
}

func (s *LoggingFlags) Set(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&s.Debug, "debug", false, "Set logging level to debug")
}
