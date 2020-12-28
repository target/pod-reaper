package main

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type replicaSafety struct {
	minimumReplicas int
	safetyMap       map[types.UID]int
}

func newReplicaSafety(minimumReplicas int, pods *v1.PodList) replicaSafety {
	safety := replicaSafety{
		minimumReplicas: minimumReplicas,
		safetyMap:       make(map[types.UID]int),
	}
	for _, pod := range pods.Items {
		for _, OwnerReference := range pod.ObjectMeta.OwnerReferences {
			safety.safetyMap[OwnerReference.UID] = safety.safetyMap[OwnerReference.UID] + 1
		}
	}
	return safety
}

func (safety *replicaSafety) isSafe(pod v1.Pod) (bool, string) {
	for _, OwnerReference := range pod.ObjectMeta.OwnerReferences {
		if safety.safetyMap[OwnerReference.UID] <= safety.minimumReplicas {
			return false, fmt.Sprintf("pod flagged as unsafe for delete by minimum replicas %s/%s", OwnerReference.Kind, OwnerReference.Name)
		}
	}
	for _, OwnerReference := range pod.ObjectMeta.OwnerReferences {
		safety.safetyMap[OwnerReference.UID] = safety.safetyMap[OwnerReference.UID] - 1
	}
	return true, ""
}
