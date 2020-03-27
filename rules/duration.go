package rules

import (
	"fmt"
	"os"
	"time"

	"k8s.io/api/core/v1"
)

const envMaxDuration = "MAX_DURATION"

var _ Rule = (*duration)(nil)

type duration struct {
	duration time.Duration
}

func (rule *duration) load() (bool, string, error) {
	value, active := os.LookupEnv(envMaxDuration)
	if !active {
		return false, "", nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return false, "", fmt.Errorf("invalid max duration: %s", err)
	}
	rule.duration = duration
	return true, fmt.Sprintf("maximum run duration %s", value), nil
}

func (rule *duration) ShouldReap(pod v1.Pod) (bool, string) {
	podStartTime := pod.Status.StartTime
	if podStartTime == nil {
		return false, ""
	}
	startTime := time.Unix(podStartTime.Unix(), 0) // convert to standard go time
	cutoffTime := time.Now().Add(-1 * rule.duration)
	runningDuration := time.Now().Sub(startTime)
	message := fmt.Sprintf("has been running for %s", runningDuration.String())
	return startTime.Before(cutoffTime), message
}
