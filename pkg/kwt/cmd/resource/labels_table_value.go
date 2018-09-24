package resource

import (
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
)

type LabelsTableValue struct {
	status map[string]interface{}
}

func NewLabelsTableValue(status map[string]interface{}) LabelsTableValue {
	return LabelsTableValue{status}
}

func (t LabelsTableValue) String() string {
	if labels, ok := t.status["labels"].(map[string]interface{}); ok {
		return uitable.NewValueInterface(labels).String()
	}
	return uitable.NewValueInterface(nil).String()
}

func (t LabelsTableValue) Value() uitable.Value { return t }

func (t LabelsTableValue) Compare(other uitable.Value) int {
	panic("Not implemented")
}
