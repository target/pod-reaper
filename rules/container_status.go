package rules

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/api/core/v1"
)

const envContainerStatus = "CONTAINER_STATUSES"

var _ Rule = (*containerStatus)(nil)

type containerStatus struct {
	reapStatuses []string
}

func (rule *containerStatus) load() (bool, string, error) {
	value, active := os.LookupEnv(envContainerStatus)
	if !active {
		return false, "", nil
	}
	rule.reapStatuses = strings.Split(value, ",")
	return true, fmt.Sprintf("container status in [%s]", value), nil
}

func (rule *containerStatus) ShouldReap(pod v1.Pod) (bool, string) {
	for _, reapStatus := range rule.reapStatuses {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			state := containerStatus.State
			// check both waiting and terminated conditions
			if (state.Waiting != nil && state.Waiting.Reason == reapStatus) ||
				(state.Terminated != nil && state.Terminated.Reason == reapStatus) {
				return true, fmt.Sprintf("has container status %s", reapStatus)
			}
		}

		// Check init containers
		for _, initContainerStatus := range pod.Status.InitContainerStatuses {
			state := initContainerStatus.State
			// Check both waiting and terminated conditions for init containers
			if (state.Waiting != nil && state.Waiting.Reason == reapStatus) ||
				(state.Terminated != nil && state.Terminated.Reason == reapStatus) {
				return true, fmt.Sprintf("has init container status %s", reapStatus)
			}
		}
	}
	return false, ""
}
