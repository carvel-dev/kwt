package workspace

import (
	cmdcore "github.com/k14s/kwt/pkg/kwt/cmd/core"
	"github.com/spf13/cobra"
)

type RunFlags struct {
	SyncFlags

	WorkingDir          string
	WorkingDirFromInput string

	Command string

	Enter bool
}

func (s *RunFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	s.SyncFlags.Set(cmd, flagsFactory)

	cmd.Flags().StringVarP(&s.WorkingDir, "directory", "d", "", "Set working directory for executing script (relative to workspace directory unless is an absolute path)")
	cmd.Flags().StringVar(&s.WorkingDirFromInput, "di", "", "Set working directory for executing script to particular's input directory")

	cmd.Flags().StringVarP(&s.Command, "command", "c", "", "Set command")

	cmd.Flags().BoolVar(&s.Enter, "enter", false, "Enter workspace after create")
}

func (s *RunFlags) HasCommand() bool {
	return len(s.Command) > 0
}
