package main

import (
	"time"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	k8v1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"

	"github.com/target/pod-reaper/rules"
)

type reaper struct {
	clientSet *kubernetes.Clientset
	options   options
}

func newReaper() reaper {
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.WithError(err).Panic("error getting in cluster kubernetes config")
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.WithError(err).Panic("unable to get client set for in cluster kubernetes config")
		panic(err)
	}
	if clientSet == nil {
		message := "kubernetes client set cannot be nil"
		logrus.Panic(message)
		panic(message)
	}
	options, err := loadOptions()
	if err != nil {
		logrus.WithError(err).Panic("error loading options")
		panic(err)
	}
	return reaper{
		clientSet: clientSet,
		options:   options,
	}
}

func (reaper reaper) getPods() *k8v1.PodList {
	coreClient := reaper.clientSet.CoreV1()
	pods := coreClient.Pods(reaper.options.namespace)
	listOptions := k8v1.ListOptions{}
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
		logrus.WithError(err).Panic("unable to get pods from the cluster")
		panic(err)
	}
	return podList
}

func (reaper reaper) reapPod(pod k8v1.Pod, reasons []string) {
	logrus.WithFields(logrus.Fields{
		"pod":     pod.Name,
		"reasons": reasons,
	}).Info("reaping pod")
	deleteOptions := &k8v1.DeleteOptions{
		GracePeriodSeconds: reaper.options.gracePeriod,
	}
	err := reaper.clientSet.CoreV1().Pods(pod.Namespace).Delete(pod.Name, deleteOptions)
	if err != nil {
		// log the error, but continue on
		logrus.WithFields(logrus.Fields{
			"pod": pod.Name,
		}).WithError(err).Warn("unable to delete pod", err)
	}
}

func (reaper reaper) scytheCycle() {
	logrus.Debug("starting reap cycle")
	pods := reaper.getPods()
	for _, pod := range pods.Items {
		shouldReap, reapReasons, spareReasons := rules.ShouldReap(pod)
		if shouldReap {
			reaper.reapPod(pod, reapReasons)
		} else if len(spareReasons) > 0 {
			// if there are explict reasons to spare the pod, log them
			logrus.WithFields(logrus.Fields{
				"pod":     pod.Name,
				"reasons": spareReasons,
			}).Debug("sparing pod")
		}
	}
	logrus.Debug("reap cycle completed")
}

func (reaper reaper) harvest() {
	runForever := reaper.options.runDuration == 0
	schedule := cron.New()
	err := schedule.AddFunc(reaper.options.schedule, func() {
		reaper.scytheCycle()
	})

	if err != nil {
		logrus.WithError(err).Panic("unable to create cron schedule: " + reaper.options.schedule)
		panic(err)
	}

	schedule.Start()

	if runForever {
		select {} // should only fail if no routine can make progress
	} else {
		time.Sleep(reaper.options.runDuration)
		schedule.Stop()
	}
}
