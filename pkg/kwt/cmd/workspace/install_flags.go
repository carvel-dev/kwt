package workspace

import (
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	"github.com/spf13/cobra"
)

type InstallFlags struct {
	Desktop      bool
	Firefox      bool
	SublimeText  bool
	GoogleChrome bool
	Go1x         bool
}

func (s *InstallFlags) Set(prefix string, cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	if len(prefix) > 0 {
		prefix += "-"
	}

	cmd.Flags().BoolVar(&s.Desktop, prefix+"desktop", false, "Configure X11 and VNC access")
	cmd.Flags().BoolVar(&s.Firefox, prefix+"firefox", false, "Install Firefox")
	cmd.Flags().BoolVar(&s.SublimeText, prefix+"sublime", false, "Install Sublime Text")
	cmd.Flags().BoolVar(&s.GoogleChrome, prefix+"chrome", false, "Install Google Chrome")
	cmd.Flags().BoolVar(&s.Go1x, prefix+"go1x", false, "Install Go 1.x")
}
