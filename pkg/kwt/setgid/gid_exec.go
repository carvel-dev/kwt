package setgid

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"syscall"
)

const (
	gidExecEnvVar    = "KWT_GIT_EXEC_GID"
	gidExecGroupName = "daemon"
)

// GidExec reexecs same process with specified GID
// because there is no reliable way of setting gid via Go
type GidExec struct{}

func (e GidExec) SetProcessGID() (int, error) {
	if os.Getenv(gidExecEnvVar) != "" {
		gidInt, err := e.gidInt(os.Getenv(gidExecEnvVar))
		if err != nil {
			return -1, err
		}
		return gidInt, e.verifyGid(gidInt)
	}

	return -1, e.execWithGid(gidExecGroupName) // 'nobody' has -2 as gid?
}

func (GidExec) gidInt(str string) (int, error) {
	grp, err := user.LookupGroup(str)
	if err != nil {
		return -1, fmt.Errorf("Looking up group '%s': %s", str, err)
	}

	gidInt, err := strconv.Atoi(grp.Gid)
	if err != nil {
		return -1, fmt.Errorf("Converting GID to int '%s': %s", grp.Gid, err)
	}

	return gidInt, nil
}

func (e GidExec) execWithGid(gidName string) error {
	gidInt, err := e.gidInt(gidName)
	if err != nil {
		return err
	}

	runtime.LockOSThread()

	err = Setgid(gidInt)
	if err != nil {
		return fmt.Errorf("Calling setgid: %s", err)
	}

	binaryPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return fmt.Errorf("Looking up binary '%s': %s", os.Args[0], err)
	}

	err = syscall.Exec(binaryPath, os.Args, append(os.Environ(), gidExecEnvVar+"="+gidName))
	if err != nil {
		return fmt.Errorf("Calling exec '%s': %s", os.Args[0], err)
	}

	panic("Unreachable")
}

func (GidExec) verifyGid(expectedGid int) error {
	actualGid := syscall.Getegid()
	if actualGid != expectedGid {
		return fmt.Errorf("Expected effective gid to be '%d' but was '%d'", expectedGid, actualGid)
	}
	return nil
}
