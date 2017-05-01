package rules

import (
	"fmt"
	"k8s.io/client-go/pkg/api/v1"
	"os"
	"strings"
)

const ENV_CONTAINER_STATUS = "CONTAINER_STATUSES"

type containerStatus struct {
	reapStatuses []string
}

func (rule *containerStatus) load() (bool, error) {
	value, active := os.LookupEnv(ENV_CONTAINER_STATUS)
	if !active {
		return false, nil
	}
	statuses := strings.Split(value, ",")
	rule.reapStatuses = statuses
	return true, nil
}

func (rule *containerStatus) ShouldReap(pod v1.Pod) (bool, string) {
	for _, reapStatus := range rule.reapStatuses {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			state := containerStatus.State
			// check both waiting and terminated conditions
			if (state.Waiting != nil && state.Waiting.Reason == reapStatus) ||
				(state.Terminated != nil && state.Terminated.Reason == reapStatus) {
				return true, fmt.Sprintf("has status of %s", reapStatus)
			}
		}
	}
	return false, ""
}
