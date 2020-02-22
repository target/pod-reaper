package rules

import (
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/pkg/api/v1"
)

const envMaxUnready = "MAX_UNREADY"

func unready(pod v1.Pod) (result, string) {
	value, active := os.LookupEnv(envMaxUnready)
	if !active {
		return ignore, notConfigured
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Errorf("failed to parse %s=%s %v", envMaxUnready, value, err))
	}
	readyCondition := getCondition(pod, v1.PodReady)
	if readyCondition == nil {
		return spare, "pod does not have a ready condition"
	}
	if readyCondition.Status == "True" {
		return spare, "pod is ready"
	}
	transitionTime := time.Unix(readyCondition.LastTransitionTime.Unix(), 0) // convert to standard go time
	cutoffTime := time.Now().Add(-1 * duration)
	unreadyDuration := time.Now().Sub(transitionTime)
	if transitionTime.Before(cutoffTime) {
		return reap, fmt.Sprintf("has been unready longer than %s (%s)", duration, unreadyDuration)
	}
	return spare, fmt.Sprintf("has been unready less than %s (%s)", duration, unreadyDuration)
}

func getCondition(pod v1.Pod, conditionType v1.PodConditionType) *v1.PodCondition {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}
