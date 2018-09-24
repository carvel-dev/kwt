package resource

import (
	"fmt"
	"strings"
	"time"

	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"k8s.io/apimachinery/pkg/util/duration"
)

type ShortConditionsTableValue struct {
	status map[string]interface{}
}

func NewShortConditionsTableValue(status map[string]interface{}) ShortConditionsTableValue {
	return ShortConditionsTableValue{status}
}

func (t ShortConditionsTableValue) NeedsAttention() bool {
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

func (t ShortConditionsTableValue) String() string {
	var result []string

	if conditions, ok := t.status["conditions"].([]interface{}); ok {
		for _, cond := range conditions {
			if typedCond, ok := cond.(map[string]interface{}); ok {
				agoStr := "?"
				if ltt, found := typedCond["lastTransitionTime"]; found {
					t, err := time.Parse(time.RFC3339, ltt.(string))
					if err == nil {
						agoStr = duration.ShortHumanDuration(time.Now().Sub(t))
					}
				}

				if ttype, found := typedCond["type"]; found {
					result = append(result, fmt.Sprintf("%s=%s(%s)", ttype, typedCond["status"], agoStr))
				}

				if state, found := typedCond["state"]; found {
					result = append(result, fmt.Sprintf("%s=%s(%s)", state, typedCond["status"], agoStr))
				}
			}
		}
	}

	return strings.Join(result, ",")
}

func (t ShortConditionsTableValue) Value() uitable.Value { return t }

func (t ShortConditionsTableValue) Compare(other uitable.Value) int {
	panic("Not implemented")
}
