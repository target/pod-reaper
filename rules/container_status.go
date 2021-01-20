package rules

import (
	"fmt"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"
)

const (
	containerStatusName       = "container_status"
	envContainerStatus        = "CONTAINER_STATUSES"
	annotationContainerStatus = annotationPrefix + "/container-statuses"
)

var _ Rule = (*containerStatus)(nil)

type containerStatus struct {
	reapStatuses []string
}

func (rule *containerStatus) load() (bool, string, error) {
	explicit := explicitLoad(containerStatusName)
	value, hasDefault := os.LookupEnv(envContainerStatus)
	if !explicit && !hasDefault {
		return false, "", nil
	}
	if value != "" {
		rule.reapStatuses = strings.Split(value, ",")
	}

	if len(rule.reapStatuses) != 0 {
		return true, fmt.Sprintf("container status in [%s]", value), nil
	}
	return true, "container status loaded explicitly", nil
}

func (rule *containerStatus) ShouldReap(pod v1.Pod) (bool, string) {
	reapStatuses := rule.reapStatuses
	annotationValue := pod.Annotations[annotationContainerStatus]
	if annotationValue != "" {
		annotationValues := strings.Split(annotationValue, ",")
		reapStatuses = append(reapStatuses, annotationValues...)
	}

	for _, reapStatus := range reapStatuses {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			state := containerStatus.State
			// check both waiting and terminated conditions
			if (state.Waiting != nil && state.Waiting.Reason == reapStatus) ||
				(state.Terminated != nil && state.Terminated.Reason == reapStatus) {
				return true, fmt.Sprintf("has container status %s", reapStatus)
			}
		}
	}
	return false, ""
}
