package net

import (
	"github.com/k14s/kwt/pkg/kwt/net/dstconn"
)

type SSHEntryPoint struct {
	opts dstconn.SSHClientConnOpts
}

var _ EntryPoint = SSHEntryPoint{}

func NewSSHEntryPoint(opts dstconn.SSHClientConnOpts) SSHEntryPoint {
	return SSHEntryPoint{opts}
}

func (f SSHEntryPoint) EntryPoint() (EntryPointSession, error) {
	return SSHEntryPointSession{f.opts}, nil
}

func (f SSHEntryPoint) Delete() error { return nil }

type SSHEntryPointSession struct {
	opts dstconn.SSHClientConnOpts
}

var _ EntryPointSession = SSHEntryPointSession{}

func (s SSHEntryPointSession) Opts() dstconn.SSHClientConnOpts { return s.opts }
func (s SSHEntryPointSession) Close() error                    { return nil }
