package rules

import (
	"fmt"
	"k8s.io/client-go/pkg/api/v1"
	"time"
	"os"
)

const envMaxDuration = "MAX_DURATION"

// max duration
type duration struct {
	duration time.Duration
}

func (rule *duration) load() (bool, error) {
	value, active := os.LookupEnv(envMaxDuration)
	if !active {
		return false, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return false, fmt.Errorf("invalid max duration: %s", err)
	}
	rule.duration = duration
	return true, nil
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
