package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	ctlkube "github.com/cppforlife/kwt/pkg/kwt/kube"
	"k8s.io/client-go/rest"
)

func (w *WorkspaceImpl) Enter() error {
	path, err := exec.LookPath("kubectl") // TODO kubeconfig
	if err != nil {
		return fmt.Errorf("Looking up kubectl binary location: %s", err)
	}

	return syscall.Exec(path, []string{path, "exec", "-it", w.pod.Name, "-c", workspaceContainerName, "bash"}, os.Environ())
}

type ExecuteOpts struct {
	WorkingDir          string
	WorkingDirFromInput string

	Command     []string
	CommandArgs []string
}

func (w *WorkspaceImpl) Execute(opts ExecuteOpts, restConfig *rest.Config) error {
	executor := ctlkube.NewExec(w.pod, workspaceContainerName, w.coreClient, restConfig)

	execOpts := ctlkube.ExecuteOpts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}

	cmd := opts.Command
	cmd = append(cmd, opts.CommandArgs...)

	return executor.Execute(cmd, execOpts)
}
