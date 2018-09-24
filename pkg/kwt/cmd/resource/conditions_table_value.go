package resource

import (
	"fmt"
	"strings"
	"time"

	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"k8s.io/apimachinery/pkg/util/duration"
)

type ConditionsTableValue struct {
	status map[string]interface{}
}

func NewConditionsTableValue(status map[string]interface{}) ConditionsTableValue {
	return ConditionsTableValue{status}
}

func (t ConditionsTableValue) NeedsAttention() bool {
	if conditions, ok := t.status["conditions"].([]interface{}); ok {
		for _, cond := range conditions {
			if typedCond, ok := cond.(map[string]interface{}); ok {
				if typedStatus, ok := typedCond["status"].(string); ok {
					if typedStatus == "False" || typedStatus == "Unknown" {
						return true
					}
				}
			}
		}
	}
	return false
}

func (t ConditionsTableValue) String() string {
	var result []string

	if conditions, ok := t.status["conditions"].([]interface{}); ok {
		for _, cond := range conditions {
			if typedCond, ok := cond.(map[string]interface{}); ok {
				desc := []string{}
				if ttype, found := typedCond["type"]; found {
					desc = append(desc, fmt.Sprintf("- %s = %s", ttype, typedCond["status"]))
				}
				if state, found := typedCond["state"]; found {
					desc = append(desc, fmt.Sprintf("- %s = %s", state, typedCond["status"]))
				}
				if reason, found := typedCond["reason"]; found {
					desc = append(desc, fmt.Sprintf("    reason: %s ", reason))
				}
				if message, found := typedCond["message"]; found {
					desc = append(desc, fmt.Sprintf("    message: %s", message))
				}
				if ltt, found := typedCond["lastTransitionTime"]; found {
					t, err := time.Parse(time.RFC3339, ltt.(string))
					if err != nil {
						desc = append(desc, "    transitioned: ???")
					} else {
						desc = append(desc, fmt.Sprintf("    transitioned: %s ago", duration.ShortHumanDuration(time.Now().Sub(t))))
					}
				}

				result = append(result, strings.Join(desc, "\n"))
			}
		}
	}

	return strings.Join(result, "\n") + "\n" // todo have table add row spacer
}

func (t ConditionsTableValue) Value() uitable.Value { return t }

func (t ConditionsTableValue) Compare(other uitable.Value) int {
	panic("Not implemented")
}
