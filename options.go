package main

import (
	"fmt"
	"github.com/target/pod-reaper/rules"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"os"
	"strings"
	"time"
)

// environment variable names
const ENV_NAMESPACE = "NAMESPACE"
const ENV_POLL_INTERVAL = "POLL_INTERVAL"
const ENV_RUN_DURATION = "RUN_DURATION"
const ENV_EXCLUDE_LABEL_KEY = "EXCLUDE_LABEL_KEY"
const ENV_EXCLUDE_LABEL_VALUES = "EXCLUDE_LABEL_VALUES"
const ENV_REQUIRE_LABEL_KEY = "REQUIRE_LABEL_KEY"
const ENV_REQUIRE_LABEL_VALUES = "REQUIRE_LABEL_VALUES"

type options struct {
	namespace        string
	pollInterval     time.Duration
	runDuration      time.Duration
	labelExclusion   *labels.Requirement
	labelRequirement *labels.Requirement
	rules            rules.Rules
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
		return nil, fmt.Errorf("did not specify %s but did specify %s", ENV_EXCLUDE_LABEL_KEY, ENV_EXCLUDE_LABEL_VALUES)
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
	labelKey, labelKeyExists := os.LookupEnv(ENV_REQUIRE_LABEL_KEY)
	labelValue, labelValuesExist := os.LookupEnv(ENV_REQUIRE_LABEL_VALUES)
	if labelKeyExists && !labelValuesExist {
		return nil, fmt.Errorf("specified %s but not %s", ENV_REQUIRE_LABEL_KEY, ENV_REQUIRE_LABEL_VALUES)
	} else if !labelKeyExists && labelValuesExist {
		return nil, fmt.Errorf("did not specify %s but did specify %s", ENV_REQUIRE_LABEL_KEY, ENV_REQUIRE_LABEL_VALUES)
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
