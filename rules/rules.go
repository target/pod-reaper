package rules

import (
	v1 "k8s.io/client-go/pkg/api/v1"
)

type rule func(v1.Pod) (result, string)

// Rules is the list of all rules
var rules = []rule{}

// ShouldReap takes a pod and makes an assessment about whether or not the pod should be
// reaped based on provided reasons for the decision
func ShouldReap(pod v1.Pod) (bool, []string, []string) {
	return shouldReap(pod, rules)
}

func shouldReap(pod v1.Pod, rules []rule) (bool, []string, []string) {
	var reapReasons []string
	var spareReasons []string
	var reapPod = false
	var sparePod = false
	for _, rule := range rules {
		result, reason := rule(pod)
		switch result {
		case reap:
			reapPod = true
			reapReasons = append(reapReasons, reason)
		case spare:
			sparePod = true
			spareReasons = append(spareReasons, reason)
		case ignore:
			// do nothing
		}
	}
	// only reap if at least one rule marked for reaping, and none marked to spare
	return reapPod && !sparePod, reapReasons, spareReasons
}
