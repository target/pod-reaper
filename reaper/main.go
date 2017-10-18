package main

import (
	"fmt"
	"time"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
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
	if options.labelExclusion != nil || options.labelRequirement != nil {
		selector := labels.NewSelector()
		if options.labelExclusion != nil {
			selector = selector.Add(*options.labelExclusion)
		}
		if options.labelRequirement != nil {
			selector = selector.Add(*options.labelRequirement)
		}
		listOptions.LabelSelector = selector.String()
	}
	podList, err := pods.List(listOptions)
	if err != nil {
		panic(err)
	}
	return podList
}

func reap(clientSet *kubernetes.Clientset, pod v1.Pod, reasons []string) {
	logrus.WithFields(logrus.Fields{
		"pod":     pod.Name,
		"reasons": reasons,
	}).Info("reaping pod")
	err := clientSet.CoreV1().Pods(pod.Namespace).Delete(pod.Name, nil)
	if err != nil {
		// log the error, but continue on: often times something else has already deleted the pod
		logrus.WithFields(logrus.Fields{
			"pod":    pod.Name,
			"reason": err.Error(),
		}).Warn("failed to reap pod")
	}
}

func scytheCycle(clientSet *kubernetes.Clientset, options options) {
	logrus.Info("executing reap cycle")
	pods := getPods(clientSet, options)
	for _, pod := range pods.Items {
		shouldReap, reasons := options.rules.ShouldReap(pod)
		if shouldReap {
			reap(clientSet, pod, reasons)
		}
	}
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	clientSet := clientSet()
	options, err := loadOptions()
	if err != nil {
		panic(err)
	}
	runForever := options.runDuration == 0

	schedule := cron.New()
	err = schedule.AddFunc(options.schedule, func() {
		scytheCycle(clientSet, options)
	})

	if err != nil {
		panic(fmt.Errorf("unable to create cron schedule: '%s' %s", options.schedule, err.Error()))
	}

	schedule.Start()

	if runForever {
		select {} // should only fail if no routine can make progress
	} else {
		time.Sleep(options.runDuration)
		schedule.Stop()
	}
	logrus.Info("pod reaper is exiting")
}
