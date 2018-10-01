package net

import (
	"runtime"

	"github.com/spf13/cobra"
)

const darwinOS = "darwin"

type DNSFlags struct {
	Recursors []string
	Map       []string
	MapExecs  []string
	MDNS      bool
}

func (s *DNSFlags) Set(cmd *cobra.Command) {
	s.SetWithPrefix(cmd, "")
}

func (s *DNSFlags) SetWithPrefix(cmd *cobra.Command, prefix string) {
	if len(prefix) != 0 {
		prefix += "-"
	}

	cmd.Flags().StringSliceVarP(&s.Recursors, prefix+"recursor", "r", nil, "Recursor (can be specified multiple times)")
	cmd.Flags().StringSliceVar(&s.Map, prefix+"map", nil, "Domain to IP mapping (can be specified multiple times) (example: 'test.=127.0.0.1')")
	cmd.Flags().StringSliceVar(&s.MapExecs, prefix+"map-exec", nil, "Domain to IP mapping command to execute periodically (can be specified multiple times) (example: 'knctl dns-map')")

	// OS X needs mDNS resolver to cover .local domain
	cmd.Flags().BoolVar(&s.MDNS, prefix+"mdns", runtime.GOOS == darwinOS, "Start MDNS server")
}
