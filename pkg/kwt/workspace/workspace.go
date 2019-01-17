package workspace

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	ctlkube "github.com/cppforlife/kwt/pkg/kwt/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type WorkspaceImpl struct {
	pod        corev1.Pod
	coreClient kubernetes.Interface
}

var _ Workspace = &WorkspaceImpl{}

func (w *WorkspaceImpl) Name() string { return w.pod.Name }

func (w *WorkspaceImpl) Image() string {
	for _, cont := range w.pod.Spec.Containers {
		return cont.Image
	}
	return ""
}

func (w *WorkspaceImpl) Privileged() bool {
	for _, cont := range w.pod.Spec.Containers {
		if cont.SecurityContext != nil {
			if cont.SecurityContext.Privileged != nil && *cont.SecurityContext.Privileged {
				return true
			}
		}
	}
	return false
}

func (w *WorkspaceImpl) Ports() []string {
	var result []string
	for _, cont := range w.pod.Spec.Containers {
		for _, port := range cont.Ports {
			result = append(result, strconv.Itoa(int(port.ContainerPort))+"/"+strings.ToLower(string(port.Protocol)))
		}
	}
	return result
}

func (w *WorkspaceImpl) State() string {
	if w.pod.DeletionTimestamp != nil {
		return "Terminating"
	}
	return string(w.pod.Status.Phase)
}

func (w *WorkspaceImpl) CreationTime() time.Time { return w.pod.CreationTimestamp.Time }

func (w *WorkspaceImpl) WaitForStart(cancelCh chan struct{}) error {
	phase, err := PodStartWaiter{w.pod, w.coreClient}.WaitForStart(cancelCh)
	if err != nil {
		return fmt.Errorf("Waiting for pod to start: %s", err)
	}

	if phase != corev1.PodRunning {
		return fmt.Errorf("Expected pod phase to be running but was %s (most likely containers terminated)", phase)
	}

	return nil
}

func (w *WorkspaceImpl) Upload(input UploadInput, restConfig *rest.Config) error {
	executor := ctlkube.NewExec(w.pod, workspaceContainerName, w.coreClient, restConfig)
	remoteDirPath := input.RemotePath(ContainerEnv{}.WorkingDir())

	// TODO stop recreating directory (note that cp will not delete deleted content)
	err := executor.Execute([]string{"/bin/rm", "-rf", remoteDirPath}, ctlkube.ExecuteOpts{})
	if err != nil {
		return fmt.Errorf("Removing remote directory: %s", err)
	}

	err = executor.Execute([]string{"/bin/mkdir", "-p", remoteDirPath}, ctlkube.ExecuteOpts{})
	if err != nil {
		return fmt.Errorf("Make remote directory: %s", err)
	}

	err = ctlkube.NewDirCp(executor).Up(input.LocalPath(), remoteDirPath)
	if err != nil {
		return fmt.Errorf("Uploading files for input '%s' (%s): %s", input.Name, input.LocalPath(), err)
	}

	return nil
}

func (w *WorkspaceImpl) Download(output DownloadOutput, restConfig *rest.Config) error {
	executor := ctlkube.NewExec(w.pod, workspaceContainerName, w.coreClient, restConfig)
	remoteDirPath := output.RemotePath(ContainerEnv{}.WorkingDir())

	err := ctlkube.NewDirCp(executor).Down(output.LocalPath(), remoteDirPath)
	if err != nil {
		return fmt.Errorf("Downloading files for output '%s' (%s): %s", output.Name, output.LocalPath(), err)
	}

	return nil
}

func (w *WorkspaceImpl) Delete(wait bool) error {
	err := w.coreClient.CoreV1().Pods(w.pod.Namespace).Delete(w.pod.Name, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Deleting workspace '%s': %s", w.pod.Name, err)
	}

	if wait {
		for {
			time.Sleep(500 * time.Millisecond)

			_, getErr := w.coreClient.CoreV1().Pods(w.pod.Namespace).Get(w.pod.Name, metav1.GetOptions{})
			if getErr != nil {
				if errors.IsNotFound(getErr) {
					return nil
				}
			}
		}
	}

	return nil
}
