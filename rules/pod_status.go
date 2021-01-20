package rules

import (
	"fmt"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"
)

const (
	podStatusName       = "pod_status"
	envPodStatus        = "POD_STATUSES"
	annotationPodStatus = annotationPrefix + "/pod-statuses"
)

var _ Rule = (*podStatus)(nil)

type podStatus struct {
	reapStatuses []string
}

func (rule *podStatus) load() (bool, string, error) {
	explicit := explicitLoad(podStatusName)
	value, hasDefault := os.LookupEnv(envPodStatus)
	if !explicit && !hasDefault {
		return false, "", nil
	}
	if value != "" {
		rule.reapStatuses = strings.Split(value, ",")
	}

	if len(rule.reapStatuses) != 0 {
		return true, fmt.Sprintf("pod status in [%s]", value), nil
	}
	return true, "pod status loaded explicitly", nil
}

func (rule *podStatus) ShouldReap(pod v1.Pod) (bool, string) {
	reapStatuses := rule.reapStatuses
	annotationValue := pod.Annotations[annotationPodStatus]
	if annotationValue != "" {
		annotationValues := strings.Split(annotationValue, ",")
		reapStatuses = append(reapStatuses, annotationValues...)
	}

	status := pod.Status.Reason
	for _, reapStatus := range reapStatuses {
		if status == reapStatus {
			return true, fmt.Sprintf("has pod status %s", reapStatus)
		}
	}
	return false, ""
}
