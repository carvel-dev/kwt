package workspace

import (
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	ctlwork "github.com/cppforlife/kwt/pkg/kwt/workspace"
	"github.com/spf13/cobra"
)

type SyncFlags struct {
	Inputs  []ctlwork.UploadInput
	Outputs []ctlwork.DownloadOutput

	Watch bool
}

func (s *SyncFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	inputs := NewUploadInputFlag(&s.Inputs)
	cmd.Flags().VarP(inputs, "input", "i", "Set inputs (format: 'name=local-dir-path' or 'name=local-dir-path:remote-dir-path') (example: knctl=.)")

	outputs := NewDownloadOutputFlag(&s.Outputs)
	cmd.Flags().VarP(outputs, "output", "o", "Set outputs (format: 'name=local-dir-path' or 'name=local-dir-path:remote-dir-path') (example: knctl=.)")

	cmd.Flags().BoolVar(&s.Watch, "watch", false, "Watch and continiously sync inputs")
}
