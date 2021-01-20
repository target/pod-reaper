package rules

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

const (
	ruleDuration          = "duration"
	envMaxDuration        = "MAX_DURATION"
	annotationMaxDuration = annotationPrefix + "/max-duration"
)

var _ Rule = (*duration)(nil)

type duration struct {
	duration time.Duration
}

func (rule *duration) load() (bool, string, error) {
	explicit := explicitLoad(ruleDuration)
	value, hasDefault := os.LookupEnv(envMaxDuration)
	if !explicit && !hasDefault {
		return false, "", nil
	}
	duration, err := time.ParseDuration(value)
	if !explicit && err != nil {
		return false, "", fmt.Errorf("invalid max duration: %s", err)
	}
	rule.duration = duration

	if rule.duration != 0 {
		return true, fmt.Sprintf("maximum run duration %s", value), nil
	}
	return true, fmt.Sprint("maximum run duration loaded explicitly"), nil
}

func (rule *duration) ShouldReap(pod v1.Pod) (bool, string) {
	duration := rule.duration
	annotationValue := pod.Annotations[annotationMaxDuration]
	if annotationValue != "" {
		annotationDuration, err := time.ParseDuration(annotationValue)
		if err == nil {
			duration = annotationDuration
		} else {
			logrus.Warnf("pod %s has invalid max duration annotation: %s", pod.Name, err)
		}
	}
	if duration == 0 {
		return false, ""
	}

	podStartTime := pod.Status.StartTime
	if podStartTime == nil {
		return false, ""
	}

	startTime := time.Unix(podStartTime.Unix(), 0) // convert to standard go time
	cutoffTime := time.Now().Add(-1 * duration)
	runningDuration := time.Now().Sub(startTime)
	message := fmt.Sprintf("has been running for %s", runningDuration.String())
	return startTime.Before(cutoffTime), message
}
