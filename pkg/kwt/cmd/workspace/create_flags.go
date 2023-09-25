package workspace

import (
	cmdcore "github.com/carvel-dev/kwt/pkg/kwt/cmd/core"
	ctlwork "github.com/carvel-dev/kwt/pkg/kwt/workspace"
	"github.com/spf13/cobra"
)

type CreateFlags struct {
	cmdcore.GenerateNameFlags

	ctlwork.CreateOpts

	Remove bool
}

func (s *CreateFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	s.GenerateNameFlags.Set(cmd, flagsFactory)

	cmd.Flags().StringVar(&s.Image, "image", "", "Set image (example: nginx)")
	cmd.Flags().StringSliceVar(&s.Command, "image-command", nil, "Set command (can be set multiple times)")
	cmd.Flags().StringSliceVar(&s.CommandArgs, "image-command-arg", nil, "Set command args (can be set multiple times)")
	cmd.Flags().BoolVarP(&s.Privileged, "privileged", "p", false, "Set privilged")
	cmd.Flags().IntSliceVar(&s.Ports, "port", nil, "Set port (multiple can be specified)")
	cmd.Flags().StringVar(&s.ServiceAccountName, "service-account", "", "Set service account (could be used for pulling private images)")

	cmd.Flags().BoolVar(&s.Remove, "rm", false, "Remove workspace after execution is finished")
}
