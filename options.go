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
	fmt.Printf("using namespace \"%s\"\n", namespace)
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
	fmt.Printf("will exclude pods with label key: %s and a value in %s", excludeLabelKey, excludeLabelValues)
	return labelExclusion
}

func createRules() {
	//errs := []error{}
	//ruleMessage := "reaping when the following rules are met:\n"
	//
	//// max duration rule
	//maxDuration, err := time.ParseDuration(environment.maxDuration)
	//if err != nil {
	//	err = fmt.Errorf("invalid max duration: %s", err)
	//	errs = append(errs, err)
	//} else if maxDuration.String() != "0s" {
	//	maxDurationRule := maxDurationRule{duration: maxDuration}
	//	rules = append(rules, maxDurationRule)
	//	ruleMessage = fmt.Sprintf("%s\t- maximum pod duration > %s\n", ruleMessage, maxDuration.String())
	//}
	//
	//// container statuses rule
	//containerStatuses := strings.Split(environment.containerStatuses, ",")
	//if len(containerStatuses) != 0 {
	//	containerStatusesRule := containerStatusesRule{reapStatuses: containerStatuses}
	//	rules = append(rules, containerStatusesRule)
	//	ruleMessage = fmt.Sprintf("%s\t- container status in %s\n", ruleMessage, containerStatuses)
	//}
	//
	//// chaos chance rule
	//chaosChance, err := strconv.ParseFloat(environment.chaosChance, 64)
	//if err != nil {
	//	err = fmt.Errorf("invalid chaos chance: %s", err)
	//	errs = append(errs, err)
	//} else if chaosChance != 0.0 {
	//	chaosChanceRule := chaosRule{chance: chaosChance}
	//	rules = append(rules, chaosChanceRule)
	//	ruleMessage = fmt.Sprintf("%s\t- choas: a random number in [0.0,1.0) < %f\n", ruleMessage, chaosChance)
	//}
	//
	//// handle error cases
	//if len(rules) == 0 {
	//	err = errors.New("no rules were defined")
	//	return
	//}
	//if len(errs) > 0 {
	//	errMessage := "Errors:\n"
	//	for index, err := range errs {
	//		errMessage = fmt.Sprintf("%s\tError %d: %s\n", errMessage, index+1, err.Error())
	//	}
	//	err = errors.New(errMessage)
	//	return
	//}
	//
	//// log the rules
	//fmt.Println(ruleMessage)
}

func loadOptions() options {
	return options{
		namespace:      namespace(),
		pollInterval:   pollInterval(),
		runDuration:    runDuration(),
		labelExclusion: labelExclusion(),
		//rules:          createRules(),
	}
}
