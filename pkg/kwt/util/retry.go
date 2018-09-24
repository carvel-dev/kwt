package util

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// Retry is different from wait.Poll because
// it does not stop retrying when error is encountered
func Retry(interval, timeout time.Duration, condFunc wait.ConditionFunc) error {
	var lastErr error
	var times int

	wait.Poll(interval, timeout, func() (bool, error) {
		done, err := condFunc()
		lastErr = err
		times++
		return done, nil
	})

	if lastErr != nil {
		return fmt.Errorf("Retried %d times: %s", times, lastErr)
	}

	return nil
}
