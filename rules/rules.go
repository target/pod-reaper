package rules

import (
	"fmt"
	"k8s.io/kubernetes/pkg/api/v1"
	"math/rand"
)

type Rule interface {
	Load() (bool, interface{}, error)
	ShouldReap(pod v1.Pod) (bool, string)
}


// status
type containerStatusesRule struct {
	reapStatuses []string
}

func (status containerStatusesRule) shouldReap(pod v1.Pod) (bool, string) {
	for _, reapStatus := range status.reapStatuses {
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

// chaos
type chaosRule struct {
	chance float64
}

func (chaos chaosRule) shouldReap(pod v1.Pod) (bool, string) {
	return rand.Float64() < chaos.chance, "was flagged for chaos"
}
