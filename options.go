package main

import (
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

func namespace() string {
	return os.Getenv(ENV_NAMESPACE)
}

func envDuration(key string, defValue string) (time.Duration, error) {
	envDuration, exists := os.LookupEnv(key)
	if !exists {
		envDuration = defValue
	}
	duration, err := time.ParseDuration(envDuration)
	if err != nil {
		return duration, fmt.Errorf("invalid %s: %s", key, err)
	}
	return duration, nil
}

func pollInterval() (time.Duration, error) {
	return envDuration(ENV_POLL_INTERVAL, "1m")
}

func runDuration() (time.Duration, error) {
	return envDuration(ENV_RUN_DURATION, "0s")
}

func labelExclusion() (*labels.Requirement, error) {
	labelKey, labelKeyExists := os.LookupEnv(ENV_EXCLUDE_LABEL_KEY)
	labelValue, labelValuesExist := os.LookupEnv(ENV_EXCLUDE_LABEL_VALUES)
	if labelKeyExists && !labelValuesExist {
		return nil, fmt.Errorf("specified %s but not %s", ENV_EXCLUDE_LABEL_KEY, ENV_EXCLUDE_LABEL_VALUES)
	} else if !labelKeyExists && labelValuesExist {
		return nil, fmt.Errorf("did not specify %s but did specifiy %s", ENV_EXCLUDE_LABEL_KEY, ENV_EXCLUDE_LABEL_VALUES)
	} else if labelKeyExists && labelValuesExist {
		return nil, nil
	}
	labelValues := strings.Split(labelValue, ",")
	labelExclusion, err := labels.NewRequirement(labelKey, selection.NotIn, labelValues)
	if err != nil {
		return nil, fmt.Errorf("could not create exclusion label: %s", err)
	}
	return labelExclusion, nil
}

func loadOptions() (options, error) {
	options := options{}
	// namespace
	options.namespace = namespace()
	// poll interval
	pollInterval, err := pollInterval()
	if err != nil {
		return options, err
	}
	options.pollInterval = pollInterval
	// run duration
	runDuration, err := runDuration()
	if err != nil {
		return options, err
	}
	options.runDuration = runDuration
	// label exclusion
	labelExclusion, err := labelExclusion()
	if err != nil {
		return options, err
	}
	options.labelExclusion = labelExclusion
	// rules
	loadedRules, err := rules.LoadRules()
	if err != nil {
		return options, err
	}
	options.rules = loadedRules
	return options, nil
}
