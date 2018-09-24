package resource

import (
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
)

type StatusTableValue struct {
	status map[string]interface{}
}

func NewStatusTableValue(status map[string]interface{}) StatusTableValue {
	return StatusTableValue{status}
}

func (t StatusTableValue) String() string {
	delete(t.status, "conditions")
	return uitable.NewValueInterface(t.status).String() + "\n" // todo have table add row spacer
}

func (t StatusTableValue) Value() uitable.Value { return t }

func (t StatusTableValue) Compare(other uitable.Value) int {
	panic("Not implemented")
}
