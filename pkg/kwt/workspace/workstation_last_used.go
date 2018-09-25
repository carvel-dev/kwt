package workspace

import (
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

func (w *WorkspaceImpl) LastUsedTime() time.Time {
	if ann, found := w.pod.Annotations[workspaceLastUsedAnn]; found {
		t, err := time.Parse(time.RFC3339, ann)
		if err == nil { // Return creation time if annotation isnt set
			return t
		}
	}
	return w.pod.CreationTimestamp.Time
}

func (w *WorkspaceImpl) MarkUse() error {
	patchJSON, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				workspaceLastUsedAnn: time.Now().UTC().Format(time.RFC3339),
			},
		},
	})

	_, err = w.coreClient.CoreV1().Pods(w.pod.Namespace).Patch(w.pod.Name, types.MergePatchType, patchJSON)
	if err != nil {
		return fmt.Errorf("Marking workspace last use: %s", err)
	}

	return nil
}
