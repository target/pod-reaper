package rules

import (
	"fmt"
	"k8s.io/client-go/pkg/api/v1"
	"os"
	"strings"
)

type containerStatusesRule struct {
	reapStatuses []string
}

func (rule *containerStatusesRule) load() bool {
	value, active := os.LookupEnv("CONTAINER_STATUSES")
	if !active {
		return false
	}
	statuses := strings.Split(value, ",")
	fmt.Printf("loading rule: container statuses %s\n", statuses)
	rule.reapStatuses = statuses
	return true
}

func (rule *containerStatusesRule) ShouldReap(pod v1.Pod) (bool, string) {
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
