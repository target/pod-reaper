package main

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

func (reaper reaper) getPods() *v1.PodList {
	coreClient := reaper.clientSet.CoreV1()
	pods := coreClient.Pods(reaper.options.namespace)
	listOptions := metav1.ListOptions{}
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
	reaper.options.podSortingStrategy(podList.Items)

	if err != nil {
		logrus.WithError(err).Panic("unable to get pods from the cluster")
		panic(err)
	}
	if reaper.options.annotationRequirement != nil {
		podList.Items = filter(reaper, podList.Items...)
	}
	return podList
}

func filter(reaper reaper, pods ...v1.Pod) []v1.Pod {
	var filtered []v1.Pod
	for _, pod := range pods {
		selector := labels.Set(pod.Annotations)
		if reaper.options.annotationRequirement.Matches(selector) {
			filtered = append(filtered, pod)
		}
	}
	return filtered
}

func (reaper reaper) reapPod(pod v1.Pod, reasons []string, reapedPods int) {
	deleteOptions := &metav1.DeleteOptions{
		GracePeriodSeconds: reaper.options.gracePeriod,
	}

	podLog := logrus.WithFields(logrus.Fields{
		"pod":     pod.Name,
		"reasons": reasons,
	})

	if reaper.options.dryRun {
		podLog.Info("pod would be reaped but pod-reaper is in dry-run mode")

		return
	}

	if reaper.options.maxPods > 0 && reapedPods >= reaper.options.maxPods {
		podLog.WithFields(logrus.Fields{
			"reapedPods": reapedPods,
			"maxPods":    reaper.options.maxPods,
		}).Info("pod would be reaped but maxPods is exceeded")

		return
	}

	podLog.Info("reaping pod")
	var err error
	if reaper.options.evict {
		err = reaper.clientSet.CoreV1().Pods(pod.Namespace).Evict(&policyv1.Eviction{
			ObjectMeta:    metav1.ObjectMeta{Namespace: pod.Namespace, Name: pod.Name},
			DeleteOptions: deleteOptions,
		})
	} else {
		err = reaper.clientSet.CoreV1().Pods(pod.Namespace).Delete(pod.Name, deleteOptions)
	}
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
	reapedPods := 0
	for _, pod := range pods.Items {
		shouldReap, reasons := reaper.options.rules.ShouldReap(pod)
		if shouldReap {
			reaper.reapPod(pod, reasons, reapedPods)
			reapedPods++
		}
	}
}

func cronWithOptionalSeconds() *cron.Cron {
	return cron.New(
		cron.WithParser(
			cron.NewParser(
				// include optional seconds
				cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)))
}

func (reaper reaper) harvest() {
	runForever := reaper.options.runDuration == 0
	schedule := cronWithOptionalSeconds()
	_, err := schedule.AddFunc(reaper.options.schedule, func() {
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
