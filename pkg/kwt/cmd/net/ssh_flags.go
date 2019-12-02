package net

import (
	"github.com/spf13/cobra"
)

type SSHFlags struct {
	User       string
	Host       string
	PrivateKey string

	Image string
}

func (s *SSHFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.User, "ssh-user", "", "SSH server username")
	cmd.Flags().StringVar(&s.Host, "ssh-host", "", "SSH server address for forwarding connections (includes port)")
	cmd.Flags().StringVar(&s.PrivateKey, "ssh-private-key", "", "Private key for connecting to SSH server (PEM format)")

	cmd.Flags().StringVar(&s.Image, "ssh-image", defaultSSHImage, "Image URL to use for starting OpenSSH on K8s")
}
