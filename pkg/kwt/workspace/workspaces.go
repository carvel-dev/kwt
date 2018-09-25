package workspace

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	workspaceLabel       = "kwt.cppforlife.com/workspace"
	workspaceLastUsedAnn = "kwt.cppforlife.com/workspace-last-used"
)

type Workspaces struct {
	namespace  string
	coreClient kubernetes.Interface
}

func NewWorkspaces(namespace string, coreClient kubernetes.Interface) Workspaces {
	return Workspaces{namespace, coreClient}
}

func (w Workspaces) List() ([]Workspace, error) {
	podsList, err := w.coreClient.CoreV1().Pods(w.namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Listing pods: %s", err)
	}

	var ws []Workspace

	for _, pod := range podsList.Items {
		if _, found := pod.Labels[workspaceLabel]; found {
			ws = append(ws, &WorkspaceImpl{pod, w.coreClient})
		}
	}

	return ws, nil
}

func (w Workspaces) Find(name string) (Workspace, error) {
	pod, err := w.coreClient.CoreV1().Pods(w.namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Finding workspace '%s': %s", name, err)
	}

	return &WorkspaceImpl{*pod, w.coreClient}, nil
}
