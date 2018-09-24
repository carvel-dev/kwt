package net

import (
	"github.com/cppforlife/kwt/pkg/kwt/net/dstconn"
)

type SSHEntryPoint struct {
	opts dstconn.SSHClientConnOpts
}

var _ EntryPoint = SSHEntryPoint{}

func NewSSHEntryPoint(opts dstconn.SSHClientConnOpts) SSHEntryPoint {
	return SSHEntryPoint{opts}
}

func (f SSHEntryPoint) EntryPoint() (dstconn.SSHClientConnOpts, error) {
	return f.opts, nil
}

func (f SSHEntryPoint) Delete() error { return nil }
