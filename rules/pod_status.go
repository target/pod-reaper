package rules

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/api/core/v1"
)

const envPodStatus = "POD_STATUSES"

var _ Rule = (*podStatus)(nil)

type podStatus struct {
	reapStatuses []string
}

func (rule *podStatus) load() (bool, string, error) {
	value, active := os.LookupEnv(envPodStatus)
	if !active {
		return false, "", nil
	}
	rule.reapStatuses = strings.Split(value, ",")
	return true, fmt.Sprintf("pod status in [%s]", value), nil
}

func (rule *podStatus) ShouldReap(pod v1.Pod) (bool, string) {
	status := pod.Status.Reason
	for _, reapStatus := range rule.reapStatuses {
		if status == reapStatus {
			return true, fmt.Sprintf("has pod status %s", reapStatus)
		}
	}
	return false, ""
}
