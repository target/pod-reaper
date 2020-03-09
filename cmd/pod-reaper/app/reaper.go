package app

import (
	"time"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Reaper struct {
	clientSet *kubernetes.Clientset
	options   options
}

func NewReaper(kubeconfig string) Reaper {
	var config *rest.Config
	var err error

	if len(kubeconfig) != 0 {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logrus.WithError(err).Panic("unable to build config from kubeconfig file")
			panic(err.Error())
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			logrus.WithError(err).Panic("error getting in cluster kubernetes config")
			panic(err)
		}
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
	return Reaper{
		clientSet: clientSet,
		options:   options,
	}
}

func (reaper Reaper) getPods() *v1.PodList {
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
	if err != nil {
		logrus.WithError(err).Panic("unable to get pods from the cluster")
		panic(err)
	}
	return podList
}

func (reaper Reaper) reapPod(pod v1.Pod, reasons []string) {
	logrus.WithFields(logrus.Fields{
		"pod":     pod.Name,
		"reasons": reasons,
	}).Info("reaping pod")
	deleteOptions := &metav1.DeleteOptions{
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

func (reaper Reaper) scytheCycle() {
	logrus.Debug("starting reap cycle")
	pods := reaper.getPods()
	for _, pod := range pods.Items {
		shouldReap, reasons := reaper.options.rules.ShouldReap(pod)
		if shouldReap {
			reaper.reapPod(pod, reasons)
		}
	}
}

func (reaper Reaper) Harvest() {
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
