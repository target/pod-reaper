package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/target/pod-reaper/rules"
)

// environment variable names
const envNamespace = "NAMESPACE"
const envGracePeriod = "GRACE_PERIOD"
const envScheduleCron = "SCHEDULE"
const envRunDuration = "RUN_DURATION"
const envExcludeLabelKey = "EXCLUDE_LABEL_KEY"
const envExcludeLabelValues = "EXCLUDE_LABEL_VALUES"
const envRequireLabelKey = "REQUIRE_LABEL_KEY"
const envRequireLabelValues = "REQUIRE_LABEL_VALUES"
const envRequireAnnotationKey = "REQUIRE_ANNOTATION_KEY"
const envRequireAnnotationValues = "REQUIRE_ANNOTATION_VALUES"
const envDryRun = "DRY_RUN"

type options struct {
	namespace             string
	gracePeriod           *int64
	schedule              string
	runDuration           time.Duration
	labelExclusion        *labels.Requirement
	labelRequirement      *labels.Requirement
	annotationRequirement *labels.Requirement
	dryRun                bool
	rules                 rules.Rules
}

func namespace() string {
	return os.Getenv(envNamespace)
}

func gracePeriod() (*int64, error) {
	envGraceDuration, exists := os.LookupEnv(envGracePeriod)
	if !exists {
		return nil, nil
	}
	duration, err := time.ParseDuration(envGraceDuration)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %s", envGracePeriod, err)
	}
	seconds := int64(duration.Seconds())
	return &seconds, nil
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

func schedule() string {
	schedule, exists := os.LookupEnv(envScheduleCron)
	if !exists {
		schedule = "@every 1m"
	}
	return schedule
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

func annotationRequirement() (*labels.Requirement, error) {
	annotationKey, annotationKeyExists := os.LookupEnv(envRequireAnnotationKey)
	annotationValue, annotationValuesExist := os.LookupEnv(envRequireAnnotationValues)
	if annotationKeyExists && !annotationValuesExist {
		return nil, fmt.Errorf("specified %s but not %s", envRequireAnnotationKey, envRequireAnnotationValues)
	} else if !annotationKeyExists && annotationValuesExist {
		return nil, fmt.Errorf("did not specify %s but did specify %s", envRequireAnnotationKey, envRequireAnnotationValues)
	} else if !annotationKeyExists && !annotationValuesExist {
		return nil, nil
	}
	annotationValues := strings.Split(annotationValue, ",")
	annotationRequirement, err := labels.NewRequirement(annotationKey, selection.In, annotationValues)
	if err != nil {
		return nil, fmt.Errorf("could not create annotation requirement: %s", err)
	}
	return annotationRequirement, nil
}

func dryRun() (bool, error) {
	value, exists := os.LookupEnv(envDryRun)
	if !exists {
		return false, nil
	}
	return strconv.ParseBool(value)
}

func loadOptions() (options options, err error) {
	options.namespace = namespace()
	options.gracePeriod, err = gracePeriod()
	if err != nil {
		return options, err
	}
	options.schedule = schedule()
	options.runDuration, err = runDuration()
	if err != nil {
		return options, err
	}
	options.labelExclusion, err = labelExclusion()
	if err != nil {
		return options, err
	}
	options.labelRequirement, err = labelRequirement()
	if err != nil {
		return options, err
	}
	options.annotationRequirement, err = annotationRequirement()
	if err != nil {
		return options, err
	}
	options.dryRun, err = dryRun()
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
