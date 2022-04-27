package main

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"math/rand"
	"os"
	"sort"
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
const envMaxPods = "MAX_PODS"
const envPodSortingStrategy = "POD_SORTING_STRATEGY"
const envEvict = "EVICT"

type options struct {
	namespace             string
	gracePeriod           *int64
	schedule              string
	runDuration           time.Duration
	labelExclusion        *labels.Requirement
	labelRequirement      *labels.Requirement
	annotationRequirement *labels.Requirement
	dryRun                bool
	maxPods               int
	podSortingStrategy    func([]v1.Pod)
	rules                 rules.Rules
	evict                 bool
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

func maxPods() (int, error) {
	value, exists := os.LookupEnv(envMaxPods)
	if !exists {
		return 0, nil
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	if v < 0 {
		return 0, nil
	}

	return v, nil
}

func getPodDeletionCost(pod v1.Pod) int32 {
	// https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/#pod-deletion-cost
	costString, present := pod.ObjectMeta.Annotations["controller.kubernetes.io/pod-deletion-cost"]
	if !present {
		return 0
	}
	// per k8s doc: invalid values should be rejected by the API server
	cost, _ := strconv.ParseInt(costString, 10, 32)
	return int32(cost)
}

func defaultSort([]v1.Pod) {}

func randomSort(pods []v1.Pod) {
	rand.Shuffle(len(pods), func(i, j int) { pods[i], pods[j] = pods[j], pods[i] })
}

func oldestFirstSort(pods []v1.Pod) {
	sort.Slice(pods, func(i, j int) bool {
		if pods[i].Status.StartTime == nil {
			return false
		}
		if pods[j].Status.StartTime == nil {
			return true
		}
		return pods[i].Status.StartTime.Unix() < pods[j].Status.StartTime.Unix()
	})
}

func youngestFirstSort(pods []v1.Pod) {
	sort.Slice(pods, func(i, j int) bool {
		if pods[i].Status.StartTime == nil {
			return false
		}
		if pods[j].Status.StartTime == nil {
			return true
		}
		return pods[j].Status.StartTime.Unix() < pods[i].Status.StartTime.Unix()
	})
}

func podDeletionCostSort(pods []v1.Pod) {
	sort.Slice(pods, func(i, j int) bool {
		return getPodDeletionCost(pods[i]) < getPodDeletionCost(pods[j])
	})
}

func podSortingStrategy() (func([]v1.Pod), error) {
	sortingStrategy, present := os.LookupEnv(envPodSortingStrategy)
	if !present {
		return defaultSort, nil
	}
	switch sortingStrategy {
	case "random":
		return randomSort, nil
	case "oldest-first":
		return oldestFirstSort, nil
	case "youngest-first":
		return youngestFirstSort, nil
	case "pod-deletion-cost":
		return podDeletionCostSort, nil
	default:
		return nil, errors.New("unknown pod sorting strategy")
	}
}

func evict() (bool, error) {
	value, exists := os.LookupEnv(envEvict)
	if !exists {
		return false, nil
	}
	return strconv.ParseBool(value)
}

func loadOptions() (options options, err error) {
	options.namespace = namespace()
	if options.gracePeriod, err = gracePeriod(); err != nil {
		return options, err
	}
	options.schedule = schedule()
	if options.runDuration, err = runDuration(); err != nil {
		return options, err
	}
	if options.labelExclusion, err = labelExclusion(); err != nil {
		return options, err
	}
	if options.labelRequirement, err = labelRequirement(); err != nil {
		return options, err
	}
	if options.annotationRequirement, err = annotationRequirement(); err != nil {
		return options, err
	}
	if options.dryRun, err = dryRun(); err != nil {
		return options, err
	}
	if options.maxPods, err = maxPods(); err != nil {
		return options, err
	}
	if options.podSortingStrategy, err = podSortingStrategy(); err != nil {
		return options, err
	}
	if options.evict, err = evict(); err != nil {
		return options, err
	}

	// rules
	if options.rules, err = rules.LoadRules(); err != nil {
		return options, err
	}
	return options, nil
}
