package main

import (
	"fmt"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/rest"
	"os"
	"strings"
	"time"
)

func maxAge() time.Duration {
	duration, err := time.ParseDuration(os.Getenv("MAX_DURATION"))
	if err != nil {
		panic(err.Error())
	}
	return duration
}

func namespace() string {
	return os.Getenv("NAMESPACE")
}

func reapPhase(phase v1.PodPhase) bool {
	reapPhases := strings.Split(os.Getenv("REAP_PODS_IN_PHASES"), ",")
	for _, reapPhase := range reapPhases {
		if phase == v1.PodPhase(reapPhase) {
			return true
		}
	}
	return false
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
	pods, err := clientSet.Core().Pods(namespace()).List(api.ListOptions{})
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
		if reapPhase(status.Phase) {
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
		time.Sleep(1 * time.Minute)
	}

}
