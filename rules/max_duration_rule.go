package rules

import (
	"fmt"
	"k8s.io/client-go/pkg/api/v1"
	"time"
)

// max duration
type maxDurationRule struct {
	duration time.Duration
}

func load() (bool, interface{}, error) {
	return false, maxDurationRule{}, nil
}

func (rule maxDurationRule) shouldReap(pod v1.Pod) (bool, string) {
	podStartTime := pod.Status.StartTime
	if podStartTime == nil {
		return false, ""
	}
	startTime := time.Unix(podStartTime.Unix(), 0) // convert to standard go time
	cutoffTime := time.Now().Add(-1 * rule.duration)
	message := fmt.Sprintf("exceeded maximum duration of %s", rule.duration.String())
	return startTime.Before(cutoffTime), message
}
