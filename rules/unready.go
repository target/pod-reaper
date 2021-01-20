package rules

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

const (
	ruleUnready          = "unready"
	envMaxUnready        = "MAX_UNREADY"
	annotationMaxUnready = annotationPrefix + "/max-unready"
)

var _ Rule = (*unready)(nil)

type unready struct {
	duration time.Duration
}

func (rule *unready) load() (bool, string, error) {
	explicit := explicitLoad(ruleUnready)
	value, hasDefault := os.LookupEnv(envMaxUnready)
	if !explicit && !hasDefault {
		return false, "", nil
	}
	duration, err := time.ParseDuration(value)
	if !explicit && err != nil {
		return false, "", fmt.Errorf("invalid max unready duration: %s", err)
	}
	rule.duration = duration

	if rule.duration != 0 {
		return true, fmt.Sprintf("maximum unready %s", value), nil
	}
	return true, fmt.Sprint("maximum unready duration loaded explicitly"), nil
}

func (rule *unready) ShouldReap(pod v1.Pod) (bool, string) {
	duration := rule.duration
	annotationValue := pod.Annotations[annotationMaxUnready]
	if annotationValue != "" {
		annotationDuration, err := time.ParseDuration(annotationValue)
		if err == nil {
			duration = annotationDuration
		} else {
			logrus.Warnf("invalid max unready duration annotation: %s", err)
		}
	}

	condition := getCondition(pod, v1.PodReady)
	if condition == nil || condition.Status == "True" {
		return false, ""
	}

	transitionTime := time.Unix(condition.LastTransitionTime.Unix(), 0) // convert to standard go time
	cutoffTime := time.Now().Add(-1 * duration)
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
