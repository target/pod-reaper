package main

import (
	"fmt"
	"os"

	"errors"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"time"
)

type reaper struct {
	clientSet *kubernetes.Clientset
	options   options
}

func newReaper() reaper {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	if clientSet == nil {
		panic(errors.New("kubernetes client set cannot be nil"))
	}
	options, err := loadOptions()
	if err != nil {
		panic(err)
	}
	return reaper{
		clientSet: clientSet,
		options:   options,
	}
}

func (reaper reaper) getPods() *v1.PodList {
	coreClient := reaper.clientSet.CoreV1()
	pods := coreClient.Pods(reaper.options.namespace)
	listOptions := v1.ListOptions{}
	if reaper.options.labelExclusion != nil || reaper.options.labelRequirement != nil {
		selector := labels.NewSelector()
		if reaper.options.labelExclusion != nil {
			selector = selector.Add(*reaper.options.labelExclusion)
		}
		if reaper.options.labelRequirement != nil {
			selector = selector.Add(*reaper.options.labelRequirement)
		}
		listOptions.LabelSelector = selector.String()
	}
	podList, err := pods.List(listOptions)
	if err != nil {
		panic(err)
	}
	return podList
}

func (reaper reaper) reapPod(pod v1.Pod, reasons []string) {
	logrus.WithFields(logrus.Fields{
		"pod":     pod.Name,
		"reasons": reasons,
	}).Info("reaping pod")
	deleteOptions := &v1.DeleteOptions{
		GracePeriodSeconds: reaper.options.gracePeriod,
	}
	err := reaper.clientSet.CoreV1().Pods(pod.Namespace).Delete(pod.Name, deleteOptions)
	if err != nil {
		// log the error, but continue on
		fmt.Fprintf(os.Stderr, "unable to delete pod %s: %s", pod.Name, err)
	}
}

func (reaper reaper) scytheCycle() {
	pods := reaper.getPods()
	for _, pod := range pods.Items {
		shouldReap, reason := reaper.options.rules.ShouldReap(pod)
		if shouldReap {
			reaper.reapPod(pod, reason)
		}
	}
}

func (reaper reaper) harvest() {
	runForever := reaper.options.runDuration == 0
	schedule := cron.New()
	err := schedule.AddFunc(reaper.options.schedule, func() {
		reaper.scytheCycle()
	})

	if err != nil {
		panic(fmt.Errorf("unable to create cron schedule: '%s' %s", reaper.options.schedule, err.Error()))
	}

	schedule.Start()

	if runForever {
		select {} // should only fail if no routine can make progress
	} else {
		time.Sleep(reaper.options.runDuration)
		schedule.Stop()
	}
}
