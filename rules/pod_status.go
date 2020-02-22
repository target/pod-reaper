package rules

import (
	"fmt"
	"os"
	"strings"

	v1 "k8s.io/client-go/pkg/api/v1"
)

const envPodStatus = "POD_STATUSES"

func podStatus(pod v1.Pod) (result, string) {
	value, active := os.LookupEnv(envPodStatus)
	if !active {
		return ignore, notConfigured
	}
	reapStatuses := strings.Split(value, ",")
	status := pod.Status.Reason
	for _, reapStatus := range reapStatuses {
		if status == reapStatus {
			return reap, fmt.Sprintf("has pod status '%s' in {%s}", reapStatus, value)
		}
	}
	return spare, fmt.Sprintf("has pod status '%s' not in {%s}", status, value)
}
