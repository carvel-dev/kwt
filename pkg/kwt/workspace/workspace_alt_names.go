package workspace

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

const (
	altNameSeparator = ","
)

func (w *WorkspaceImpl) AltNames() []string {
	if ann, found := w.pod.Annotations[workspaceAltNamesAnn]; found {
		return strings.Split(ann, altNameSeparator)
	}
	return nil
}

func (w *WorkspaceImpl) AddAltName(name string) error {
	existingNames := w.AltNames()
	existingNames = append(existingNames, name)

	patchJSON, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				workspaceAltNamesAnn: strings.Join(existingNames, altNameSeparator),
			},
		},
	})

	_, err = w.coreClient.CoreV1().Pods(w.pod.Namespace).Patch(w.pod.Name, types.MergePatchType, patchJSON)
	if err != nil {
		return fmt.Errorf("Updating workspace additional names: %s", err)
	}

	return nil
}
