package rules

import (
	"fmt"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"
)

const envPodStatusPhase = "POD_STATUS_PHASES"

var _ Rule = (*podStatusPhase)(nil)

type podStatusPhase struct {
	reapStatusPhases []string
}

func (rule *podStatusPhase) load() (bool, string, error) {
	value, active := os.LookupEnv(envPodStatusPhase)
	if !active {
		return false, "", nil
	}
	rule.reapStatusPhases = strings.Split(value, ",")
	return true, fmt.Sprintf("pod status phase in [%s]", value), nil
}

func (rule *podStatusPhase) ShouldReap(pod v1.Pod) (bool, string) {
	status := string(pod.Status.Phase)
	for _, reapStatusPhase := range rule.reapStatusPhases {
		if status == reapStatusPhase {
			return true, fmt.Sprintf("has pod status phase %s", reapStatusPhase)
		}
	}
	return false, ""
}
