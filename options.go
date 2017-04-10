package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Options struct {
	maxPodDuration    time.Duration
	pollInterval      time.Duration
	containerStatuses []string
	excludeLabelKey   string
	excludeLabelValue string
	namespace         string
}

func environment(environmentVariable string, defaultValue string) string {
	env := os.Getenv(environmentVariable)
	if env == "" {
		return defaultValue
	}
	return env
}

func duration(environmentVariable string, defaultValue string) time.Duration {
	env := environment(environmentVariable, defaultValue)
	duration, err := time.ParseDuration(env)
	if err != nil {
		panic(err)
	}
	return duration
}

func split(environmentVariable string) []string {
	env := os.Getenv(environmentVariable)
	if env == "" {
		return []string{}
	}
	return strings.Split(env, ",")
}

func options() Options {
	return Options{
		maxPodDuration:    duration("MAX_POD_DURATION", "2h"),
		pollInterval:      duration("POLL_INTERVAL", "15s"),
		containerStatuses: split("CONTAINER_STATUSES"),
		excludeLabelKey:   environment("EXCLUDE_LABEL_KEY", "pod-reaper"),
		excludeLabelValue: environment("EXCLUDE_LABEL_VALUE", "disabled"),
		namespace:         environment("NAMESPACE", ""),
	}
}

func (options *Options) printOptions() {
	// write out the used environment variables
	fmt.Printf("MAX_POD_DURATION: %s\n", options.maxPodDuration)
	fmt.Printf("POLL_INTERVAL: %s\n", options.pollInterval)
	fmt.Printf("CONTAINER_STATUSES: %s\n", strings.Join(options.containerStatuses, ", "))
	fmt.Printf("EXCLUDE_LABEL_KEY: %s\n", options.excludeLabelKey)
	fmt.Printf("EXCLUDE_LABEL_VALUE: %s\n", options.excludeLabelValue)
	fmt.Printf("NAMESPACE: %s\n", options.namespace)
}
