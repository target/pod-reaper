package rules

import (
	"fmt"
	"k8s.io/client-go/pkg/api/v1"
	"time"
	"os"
)

// max duration
type maxDurationRule struct {
	duration time.Duration
}

func (rule *maxDurationRule) load() bool {
	value, active := os.LookupEnv("MAX_DURATION")
	if !active {
		return false
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Errorf("invalid max duration: %s", err))
	}
	fmt.Printf("loading rule: max pod duration %s\n", duration.String())
	rule.duration = duration
	return true
}

func (rule *maxDurationRule) ShouldReap(pod v1.Pod) (bool, string) {
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
