package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/target/pod-reaper/rules"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// environment variable names
const envNamespace = "NAMESPACE"
const envPollInterval = "POLL_INTERVAL"
const envRunDuration = "RUN_DURATION"
const envExcludeLabelKey = "EXCLUDE_LABEL_KEY"
const envExcludeLabelValues = "EXCLUDE_LABEL_VALUES"
const envRequireLabelKey = "REQUIRE_LABEL_KEY"
const envRequireLabelValues = "REQUIRE_LABEL_VALUES"

type options struct {
	namespace        string
	pollInterval     time.Duration
	runDuration      time.Duration
	labelExclusion   *labels.Requirement
	labelRequirement *labels.Requirement
	rules            rules.Rules
}

func namespace() string {
	return os.Getenv(envNamespace)
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
	return envDuration(envPollInterval, "1m")
}

func runDuration() (time.Duration, error) {
	return envDuration(envRunDuration, "0s")
}

func labelExclusion() (*labels.Requirement, error) {
	labelKey, labelKeyExists := os.LookupEnv(envExcludeLabelKey)
	labelValue, labelValuesExist := os.LookupEnv(envExcludeLabelValues)
	if labelKeyExists && !labelValuesExist {
		return nil, fmt.Errorf("specified %s but not %s", envExcludeLabelKey, envExcludeLabelValues)
	} else if !labelKeyExists && labelValuesExist {
		return nil, fmt.Errorf("did not specify %s but did specify %s", envExcludeLabelKey, envExcludeLabelValues)
	} else if !labelKeyExists && !labelValuesExist {
		return nil, nil
	}
	labelValues := strings.Split(labelValue, ",")
	labelExclusion, err := labels.NewRequirement(labelKey, selection.NotIn, labelValues)
	if err != nil {
		return nil, fmt.Errorf("could not create exclusion label: %s", err)
	}
	return labelExclusion, nil
}

func labelRequirement() (*labels.Requirement, error) {
	labelKey, labelKeyExists := os.LookupEnv(envRequireLabelKey)
	labelValue, labelValuesExist := os.LookupEnv(envRequireLabelValues)
	if labelKeyExists && !labelValuesExist {
		return nil, fmt.Errorf("specified %s but not %s", envRequireLabelKey, envRequireLabelValues)
	} else if !labelKeyExists && labelValuesExist {
		return nil, fmt.Errorf("did not specify %s but did specify %s", envRequireLabelKey, envRequireLabelValues)
	} else if !labelKeyExists && !labelValuesExist {
		return nil, nil
	}
	labelValues := strings.Split(labelValue, ",")
	labelRequirement, err := labels.NewRequirement(labelKey, selection.In, labelValues)
	if err != nil {
		return nil, fmt.Errorf("could not create requirement label: %s", err)
	}
	return labelRequirement, nil
}

func loadOptions() (options options, err error) {
	// namespace
	options.namespace = namespace()
	// poll interval
	options.pollInterval, err = pollInterval()
	if err != nil {
		return options, err
	}
	// run duration
	options.runDuration, err = runDuration()
	if err != nil {
		return options, err
	}
	// label exclusion
	options.labelExclusion, err = labelExclusion()
	if err != nil {
		return options, err
	}
	// label requirement
	options.labelRequirement, err = labelRequirement()
	if err != nil {
		return options, err
	}

	// rules
	options.rules, err = rules.LoadRules()
	if err != nil {
		return options, err
	}
	return options, err
}