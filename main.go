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
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientSet
}

func getPods(clientSet *kubernetes.Clientset, options options) *v1.PodList {
	coreClient := clientSet.CoreV1()
	pods := coreClient.Pods(options.namespace)
	listOptions := v1.ListOptions{}
	if options.labelExclusion != nil {
		selector := labels.NewSelector().Add(*options.labelExclusion)
		listOptions.LabelSelector = selector.String()
	}
	podList, err := pods.List(listOptions)
	if err != nil {
		panic(err)
	}
	return podList
}

func reap(clientSet *kubernetes.Clientset, pod v1.Pod, reason string) {
	fmt.Printf("Reaping Pod %s because %s\n", pod.Name, reason)
	err := clientSet.Core().Pods(pod.Namespace).Delete(pod.Name, nil)
	if err != nil {
		// log the error, but continue on
		fmt.Fprintf(os.Stderr, "unable to delete pod %s because %s", pod.Name, err)
	}
}

func scytheCycle(clientSet *kubernetes.Clientset, options options) {
	pods := getPods(clientSet, options)
	for _, pod := range pods.Items {
		shouldReap, reason := options.rules.ShouldReap(pod)
		if shouldReap {
			reap(clientSet, pod, reason)
		}
	}
}

func main() {
	clientSet := clientSet()
	options, err := loadOptions()
	if err != nil {
		panic(err)
	}
	runForever := options.runDuration == 0
	cutoff := time.Now().Add(options.runDuration)
	for {
		scytheCycle(clientSet, options)
		if !runForever && time.Now().After(cutoff) {
			os.Exit(0) // successful exit
		}
		time.Sleep(options.pollInterval)
	}
}
