package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
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

	//labelExclusion := options.labelExclusion
	//if labelExclusion == nil {
	//	pods, err := clientSet.CoreV1().Pods(options.namespace).List(v1.ListOptions{})
	//	if err != nil {
	//		panic(err.Error())
	//	}
	//	return pods
	//
	//}
	//selector := labels.NewSelector().Add(options.labelExclusion)
	//pods, err := clientSet.CoreV1().Pods(options.namespace).List(v1.ListOptions{LabelSelector:selector})
	//if err != nil {
	//	panic(err.Error())
	//}
	//return pods
}

func main() {
	options := loadOptions()
	pods := getPods(clientSet(), options)
	for _, pod := range pods.Items {
		fmt.Printf("Found pod: %s\n", pod.Name)
	}
	fmt.Println("hello")
}
