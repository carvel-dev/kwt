package setgid

import (
	"syscall"
)

func Setgid(gid int) error {
	return syscall.Setgid(gid)
}
