package main

import (
	"fmt"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/labels"
	"k8s.io/client-go/1.5/pkg/selection"
	"k8s.io/client-go/1.5/pkg/util/sets"
	"k8s.io/client-go/1.5/rest"
	"strings"
	"time"
)

func clientSet() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientSet
}

func getPods(clientSet *kubernetes.Clientset, options Options) *v1.PodList {
	excludeValueSet := sets.NewString(options.excludeLabelValue)
	requirement, err := labels.NewRequirement(options.excludeLabelKey, selection.NotEquals, excludeValueSet)
	if err != nil {
		panic(err.Error())
	}
	selector := labels.NewSelector().Add(*requirement)
	pods, err := clientSet.Core().Pods(options.namespace).List(api.ListOptions{LabelSelector: selector})
	if err != nil {
		panic(err.Error())
	}
	return pods
}

func reap(options Options) {
	clientSet := clientSet()
	pods := getPods(clientSet, options)
	for _, pod := range pods.Items {
		shouldDelete, message := shouldDelete(pod, options)
		if shouldDelete {
			fmt.Println(message)
			err := clientSet.Core().Pods(pod.ObjectMeta.Namespace).Delete(pod.ObjectMeta.Name, nil)
			if err != nil {
				panic(err)
			}
		}
	}
}

func hasStatus(pod v1.Pod, containerStates []string) (bool, string) {
	for _, reason := range containerStates {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			state := containerStatus.State
			if state.Waiting != nil && state.Waiting.Reason == reason {
				return true, reason
			}
			if state.Terminated != nil && state.Terminated.Reason == reason {
				return true, reason
			}
		}
	}
	return false, pod.Status.Reason
}

func exceedsMaxDuration(pod v1.Pod, maxPodDuration time.Duration) bool {
	cutoff := unversioned.Unix(time.Now().Add(-1*maxPodDuration).Unix(), 0)
	return pod.Status.StartTime.Before(cutoff)
}

func shouldDelete(pod v1.Pod, options Options) (bool, string) {
	reasons := []string{}
	if exceedsMaxDuration(pod, options.maxPodDuration) {
		reasons = append(reasons, fmt.Sprintf("it execeed max duration of %s", options.maxPodDuration))
	}
	hasStatus, podStatus := hasStatus(pod, options.containerStatuses)
	if hasStatus {
		reasons = append(reasons, fmt.Sprintf("it has status of %s", podStatus))
	}
	if len(reasons) > 0 {
		podName := pod.GetObjectMeta().GetName()
		message := fmt.Sprintf("Reaping pod: %s because %s", podName, strings.Join(reasons, " and "))
		return true, message
	}
	return false, ""
}

func main() {
	options := options()
	options.printOptions()
	for {
		reap(options)
		time.Sleep(options.pollInterval)
	}
}
