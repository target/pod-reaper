package rules

import (
	k8v1 "k8s.io/client-go/pkg/api/v1"
)

const notConfigured = "not configured"

type rule func(k8v1.Pod) (result, string)

// Rules is the list of all rules
var rules = []rule{
	chaos,
	containerStatus,
	duration,
	podStatus,
	unready,
}

// ShouldReap takes a pod and makes an assessment about whether or not the pod should be
// reaped based on provided reasons for the decision
func ShouldReap(pod k8v1.Pod) (bool, []string, []string) {
	return shouldReap(pod, rules)
}

func shouldReap(pod k8v1.Pod, rules []rule) (bool, []string, []string) {
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
