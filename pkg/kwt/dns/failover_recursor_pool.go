package dns

import (
	"errors"
	"sync/atomic"
)

const (
	FailHistoryLength    = 25
	FailHistoryThreshold = 5
)

type RecursorPool interface {
	PerformStrategically(func(string) error) error
}

type FailoverRecursorPool struct {
	preferredRecursorIndex uint64
	recursors              []recursorWithHistory

	logger Logger
	logTag string
}

type recursorWithHistory struct {
	name       string
	failBuffer chan bool
	failCount  int32
}

func NewFailoverRecursorPool(recursors []string, logger Logger) RecursorPool {
	logTag := "dns.FailoverRecursorPool"
	recursorsWithHistory := []recursorWithHistory{}

	for _, name := range recursors {
		failBuffer := make(chan bool, FailHistoryLength)
		for i := 0; i < FailHistoryLength; i++ {
			failBuffer <- false
		}

		recursorsWithHistory = append(recursorsWithHistory, recursorWithHistory{
			name:       name,
			failBuffer: failBuffer,
			failCount:  0,
		})
	}

	if len(recursorsWithHistory) > 0 {
		logger.Info(logTag, "Starting with '%s'", recursorsWithHistory[0].name)
	}

	return &FailoverRecursorPool{
		recursors:              recursorsWithHistory,
		preferredRecursorIndex: 0,

		logger: logger,
		logTag: logTag,
	}
}

func (q *FailoverRecursorPool) PerformStrategically(work func(string) error) error {
	offset := atomic.LoadUint64(&q.preferredRecursorIndex)
	uintRecursorCount := uint64(len(q.recursors))

	for i := uint64(0); i < uintRecursorCount; i++ {
		index := int((i + offset) % uintRecursorCount)
		err := work(q.recursors[index].name)
		if err == nil {
			q.registerResult(index, false)
			return nil
		}

		failures := q.registerResult(index, true)
		if i == 0 && failures >= FailHistoryThreshold {
			q.shiftPreference()
		}
	}

	return errors.New("No response from recursors")
}

func (q *FailoverRecursorPool) shiftPreference() {
	pri := atomic.AddUint64(&q.preferredRecursorIndex, 1)
	index := pri % uint64(len(q.recursors))
	q.logger.Info(q.logTag, "Shifting to '%s'", q.recursors[index].name)
}

func (q *FailoverRecursorPool) registerResult(index int, wasError bool) int32 {
	failingRecursor := &q.recursors[index]

	oldestResult := <-failingRecursor.failBuffer
	failingRecursor.failBuffer <- wasError

	change := int32(0)

	if oldestResult {
		change--
	}

	if wasError {
		change++
	}

	return atomic.AddInt32(&failingRecursor.failCount, change)
}
