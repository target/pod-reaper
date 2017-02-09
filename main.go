package main

import (
	"fmt"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/rest"
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

func main() {
	clientSet := clientSet()
	pods, err := clientSet.Core().Pods("").List(api.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	cutOffUnixSeconds := time.Now().Add(-3 * time.Hour).Unix()
	cutoff := unversioned.Unix(cutOffUnixSeconds, 0)
	for _, pod := range pods.Items {
		if pod.Status.StartTime.Before(cutoff) {
			err := clientSet.Core().Pods(pod.ObjectMeta.Namespace).Delete(pod.ObjectMeta.Name, nil)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}
