//go:build linux
// +build linux

package forwarder

import (
	"syscall"
)

const solIP = syscall.SOL_IP
