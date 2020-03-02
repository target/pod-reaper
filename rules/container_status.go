package rules

import (
	"fmt"
	"os"
	"strings"

	k8v1 "k8s.io/client-go/pkg/api/v1"
)

const envContainerStatus = "CONTAINER_STATUSES"

func containerStatus(pod k8v1.Pod) (result, string) {
	value, active := os.LookupEnv(envContainerStatus)
	if !active {
		return ignore, notConfigured
	}
	reapStatuses := strings.Split(value, ",")
	for _, reapStatus := range reapStatuses {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			state := containerStatus.State
			// check both waiting and terminated conditions
			if (state.Waiting != nil && state.Waiting.Reason == reapStatus) ||
				(state.Terminated != nil && state.Terminated.Reason == reapStatus) {
				return reap, fmt.Sprintf("has container with status '%s' in {%s}", reapStatus, value)
			}
		}
	}
	return spare, fmt.Sprintf("has no container with status in {%s}", value)
}
