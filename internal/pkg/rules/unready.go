package rules

import (
	"fmt"
	"os"
	"time"

	"k8s.io/api/core/v1"
)

const envMaxUnready = "MAX_UNREADY"

var _ Rule = (*unready)(nil)

type unready struct {
	duration time.Duration
}

func (rule *unready) load() (bool, string, error) {
	value, active := os.LookupEnv(envMaxUnready)
	if !active {
		return false, "", nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return false, "", fmt.Errorf("invalid max unready duration: %s", err)
	}
	rule.duration = duration
	return true, fmt.Sprintf("maximum unready %s", value), nil
}

func (rule *unready) ShouldReap(pod v1.Pod) (bool, string) {
	condition := getCondition(pod, v1.PodReady)
	if condition == nil || condition.Status == "True" {
		return false, ""
	}

	transitionTime := time.Unix(condition.LastTransitionTime.Unix(), 0) // convert to standard go time
	cutoffTime := time.Now().Add(-1 * rule.duration)
	unreadyDuration := time.Now().Sub(transitionTime)
	message := fmt.Sprintf("has been unready for %s", unreadyDuration.String())
	return transitionTime.Before(cutoffTime), message
}

func getCondition(pod v1.Pod, conditionType v1.PodConditionType) *v1.PodCondition {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}

	return nil
}
