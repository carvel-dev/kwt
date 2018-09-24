package resource

import (
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
)

type AnnotationsTableValue struct {
	status map[string]interface{}
}

func NewAnnotationsTableValue(status map[string]interface{}) AnnotationsTableValue {
	return AnnotationsTableValue{status}
}

func (t AnnotationsTableValue) String() string {
	if annotations, ok := t.status["annotations"].(map[string]interface{}); ok {
		return uitable.NewValueInterface(annotations).String()
	}
	return uitable.NewValueInterface(nil).String()
}

func (t AnnotationsTableValue) Value() uitable.Value { return t }

func (t AnnotationsTableValue) Compare(other uitable.Value) int {
	panic("Not implemented")
}
