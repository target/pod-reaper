package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"os"
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

func getPods(clientSet *kubernetes.Clientset, options options) *v1.PodList {
	coreClient := clientSet.CoreV1()
	pods := coreClient.Pods(options.namespace)

	listOptions := v1.ListOptions{}
	if options.labelExclusion != nil {
		selector := labels.NewSelector().Add(*options.labelExclusion)
		listOptions = v1.ListOptions{LabelSelector: selector.String()}
	}
	podList, err := pods.List(listOptions)
	if err != nil {
		panic(err.Error())
	}
	return podList
}

func main() {
	options := loadOptions()
	runForever := options.runDuration == 0
	cutoff := time.Now().Add(options.runDuration)
	pods := getPods(clientSet(), options)
	for {
		for _, pod := range pods.Items {
			fmt.Printf("Found pod: %s\n", pod.Name)
		}
		if !runForever && time.Now().After(cutoff) {
			fmt.Printf("Duration %s has elapsed, stopping execution\n", options.runDuration.String())
			os.Exit(0) // successful exit
		}
		time.Sleep(options.pollInterval)
	}

}
