package main

import (
	"fmt"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/labels"
	"k8s.io/client-go/1.5/pkg/selection"
	"k8s.io/client-go/1.5/pkg/util/sets"
	"k8s.io/client-go/1.5/rest"
	"os"
	"time"
)

func maxAge() time.Duration {
	duration, err := time.ParseDuration(os.Getenv("MAX_DURATION"))
	if err != nil {
		panic(err.Error())
	}
	return duration
}

func pollInterval() time.Duration {
	interval, err := time.ParseDuration(os.Getenv("POLL_INTERVAL"))
	if err != nil {
		panic(err.Error())
	}
	return interval
}

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

func reap() {
	clientSet := clientSet()
	selectorKey := os.Getenv("LABEL_KEY")
	selectorValue := sets.NewString(os.Getenv("LABEL_VALUE"))
	requirement, err := labels.NewRequirement(selectorKey, selection.NotEquals, selectorValue)
	if err != nil {
		panic(err.Error())
	}
	selector := labels.NewSelector().Add(*requirement)
	pods, err := clientSet.Core().Pods(os.Getenv("NAMESPACE")).List(api.ListOptions{LabelSelector: selector})
	if err != nil {
		panic(err.Error())
	}
	cutOffUnixSeconds := time.Now().Add(-1 * maxAge()).Unix()
	cutoff := unversioned.Unix(cutOffUnixSeconds, 0)
	for _, pod := range pods.Items {
		status := pod.Status
		if status.StartTime.Before(cutoff) {
			err := clientSet.Core().Pods(pod.ObjectMeta.Namespace).Delete(pod.ObjectMeta.Name, nil)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}

func main() {
	for {
		reap()
		time.Sleep(pollInterval())
	}

}
