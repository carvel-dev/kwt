package workspace

import (
	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	"github.com/spf13/cobra"
)

type DeleteFlags struct {
	Wait bool
}

func (s *DeleteFlags) Set(prefix string, cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	if len(prefix) > 0 {
		prefix += "-"
	}

	cmd.Flags().BoolVar(&s.Wait, prefix+"wait", false, "Wait for deletion to complete")
}
