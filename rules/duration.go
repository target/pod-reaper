package rules

import (
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/pkg/api/v1"
)

const envMaxDuration = "MAX_DURATION"

func duration(pod v1.Pod) (result, string) {
	value, active := os.LookupEnv(envMaxDuration)
	if !active {
		return ignore, notConfigured
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Errorf("failed to parse %s=%s %v", envMaxDuration, value, err))
	}
	podStartTime := pod.Status.StartTime
	if podStartTime == nil {
		return ignore, "pod has no start time"
	}
	startTime := time.Unix(podStartTime.Unix(), 0) // convert to standard go time
	cutoffTime := time.Now().Add(-1 * duration)
	running := time.Now().Sub(startTime)
	if startTime.Before(cutoffTime) {
		return reap, fmt.Sprintf("pod running for longer than %s (%s)", duration.String(), running.String())
	}
	return spare, fmt.Sprintf("pod running for less than %s (%s)", duration.String(), running.String())
}
