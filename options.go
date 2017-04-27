package main

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"os"
	"pod-reaper/rules"
	"strings"
	"time"
)

// environment variable names
const ENV_NAMESPACE = "NAMESPACE"
const ENV_POLL_INTERVAL = "POLL_INTERVAL"
const ENV_RUN_DURATION = "RUN_DURATION"
const ENV_EXCLUDE_LABEL_KEY = "EXCLUDE_LABEL_KEY"
const ENV_EXCLUDE_LABEL_VALUES = "EXCLUDE_LABEL_VALUES"

type options struct {
	namespace      string
	pollInterval   time.Duration
	runDuration    time.Duration
	labelExclusion *labels.Requirement
	rules          []rules.Rule
}

func environmentVariable(key string, defValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defValue
	}
	return value
}
func environmentVariableSlice(key string) []string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return []string{}
	}
	return strings.Split(value, ",")
}

func namespace() (namespace string) {
	namespace = environmentVariable(ENV_NAMESPACE, "")
	if namespace == "" {
		fmt.Println("using all namespaces (used if namespace is set to \"\")")
	} else {
		fmt.Printf("using namespace \"%s\"\n", namespace)
	}
	return
}

func pollInterval() time.Duration {
	duration, err := time.ParseDuration(environmentVariable(ENV_POLL_INTERVAL, "1m"))
	if err != nil {
		panic(fmt.Errorf("invalid poll interval: %s", err))
	}
	fmt.Printf("using poll interval \"%s\"\n", duration.String())
	return duration
}

func runDuration() time.Duration {
	duration, err := time.ParseDuration(environmentVariable(ENV_RUN_DURATION, "0s"))
	if err != nil {
		panic(fmt.Errorf("invalid run duration: %s", err))
	}
	if duration == 0 {
		fmt.Println("using indefinite run duration (used if run duration is specified to 0s)")
	} else {
		fmt.Printf("using run duration \"%s\"\n", duration.String())
	}
	return duration
}

func labelExclusion() *labels.Requirement {
	excludeLabelKey := environmentVariable(ENV_EXCLUDE_LABEL_KEY, "")
	excludeLabelValues := environmentVariableSlice(ENV_EXCLUDE_LABEL_VALUES)
	if excludeLabelKey == "" && len(excludeLabelValues) != 0 {
		panic(errors.New("specified exclude label values but did not specifiy exclude label key"))
	} else if excludeLabelKey != "" && len(excludeLabelValues) == 0 {
		panic(errors.New("specified exclude label key but did not specify exclude label values"))
	} else if excludeLabelKey == "" && len(excludeLabelValues) == 0 {
		return nil
	}
	labelExclusion, err := labels.NewRequirement(excludeLabelKey, selection.NotIn, excludeLabelValues)
	if err != nil {
		panic(fmt.Errorf("could not create exclude label: %s", err))
	}
	fmt.Printf("using label exclusion, ignoring pods with label key: %s and value in %s\n", excludeLabelKey, excludeLabelValues)
	return labelExclusion
}

func loadOptions() options {
	return options{
		namespace:      namespace(),
		pollInterval:   pollInterval(),
		runDuration:    runDuration(),
		labelExclusion: labelExclusion(),
		rules:          rules.LoadRules(),
	}
}
