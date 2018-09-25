package workspace

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	ctlkube "github.com/cppforlife/kwt/pkg/kwt/kube"
	corev1 "k8s.io/api/core/v1"
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

func (w *WorkspaceImpl) Ports() []string {
	var result []string
	for _, cont := range w.pod.Spec.Containers {
		for _, port := range cont.Ports {
			result = append(result, strconv.Itoa(int(port.ContainerPort))+"/"+strings.ToLower(string(port.Protocol)))
		}
	}
	return result
}

func (w *WorkspaceImpl) CreationTime() time.Time { return w.pod.CreationTimestamp.Time }

func (w *WorkspaceImpl) WaitForStart(cancelCh chan struct{}) error {
	_, err := PodStartWaiter{w.pod, w.coreClient}.WaitForStart(cancelCh)
	if err != nil {
		return fmt.Errorf("Waiting for pod to start: %s", err)
	}

	return nil
}

func (w *WorkspaceImpl) Upload(input UploadInput, restConfig *rest.Config) error {
	remoteDirPath := input.RemotePath(ContainerEnv{}.WorkingDir())

	executor := ctlkube.NewExec(w.pod, workspaceDebugContainerName, w.coreClient, restConfig)

	err := executor.Execute([]string{"/bin/rm", "-rf", remoteDirPath}, ctlkube.ExecuteOpts{})
	if err != nil {
		return fmt.Errorf("Removing remote directory: %s", err)
	}

	err = executor.Execute([]string{"/bin/mkdir", "-p", remoteDirPath}, ctlkube.ExecuteOpts{})
	if err != nil {
		return fmt.Errorf("Make remote directory: %s", err)
	}

	// TODO does not remove removed directories
	err = ctlkube.NewDirCp(executor).Execute(input.LocalDirPath, remoteDirPath)
	if err != nil {
		return fmt.Errorf("Uploading files for '%s' (%s): %s", input.Name, input.LocalDirPath, err)
	}

	return nil
}

func (w *WorkspaceImpl) Download(output DownloadOutput, restConfig *rest.Config) error {
	return nil
}

func (w *WorkspaceImpl) Delete() error {
	err := w.coreClient.CoreV1().Pods(w.pod.Namespace).Delete(w.pod.Name, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Deleting workspace '%s': %s", w.pod.Name, err)
	}

	return nil
}
