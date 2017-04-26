package main

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"pod-reaper/rules"
	"strings"
	"time"
)

type options struct {
	namespace      string
	pollInterval   time.Duration
	runDuration    time.Duration
	labelExclusion *labels.Requirement
	rules          []rules.Rule
}

func namespace(environment configuration) (namespace string) {
	namespace = environment.namespace
	fmt.Printf("using namespace \"%s\"\n", namespace)
	return
}

func pollInterval(environment configuration) (duration time.Duration, err error) {
	duration, err = time.ParseDuration(environment.pollInterval)
	if err != nil {
		err = fmt.Errorf("invalid poll interval: %s", err)
		return
	}
	fmt.Printf("using poll interval \"%s\"\n", duration.String())
	return
}

func runDuration(environment configuration) (duration time.Duration, err error) {
	duration, err = time.ParseDuration(environment.runDuration)
	if err != nil {
		err = fmt.Errorf("invalid run duration: %s", err)
		return
	}
	if duration == 0 {
		fmt.Println("using indefinite run duration")
		return
	}
	fmt.Printf("using run duration \"%s\"\n", duration.String())
	return
}

func labelExclusion(environment configuration) (labelExclusion *labels.Requirement, err error) {
	excludeLabelKey := environment.excludeLabelKey
	excludeLabelValues := strings.Split(environment.excludeLabelValues, ",")
	if excludeLabelKey == "" && len(excludeLabelValues) != 0 {
		err = errors.New("specified exclude label values but did not specifiy exclude label key")
		return
	} else if excludeLabelKey != "" && len(excludeLabelValues) == 0 {
		err = errors.New("specified exclude label key but did not specify exclude label values")
		return
	} else if excludeLabelKey == "" && len(excludeLabelValues) == 0 {

		return
	}
	labelExclusion, err = labels.NewRequirement(excludeLabelKey, selection.NotIn, excludeLabelValues)
	if err != nil {
		err = fmt.Errorf("could not create exclude label: %s", err)
	}
	fmt.Printf("will exclude pods with label key: %s and a value in %s", excludeLabelKey, excludeLabelValues)
	return
}

func getRules(environment configuration) (rules []rules.Rule, err error) {
	err = nil
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
	return
}

func loadOptions() options {
	errs := []error{}
	configuration := loadConfiguration()
	namespace := namespace(configuration)
	pollInterval, err := pollInterval(configuration)
	if err != nil {
		errs = append(errs, err)
	}
	runDuration, err := runDuration(configuration)
	if err != nil {
		errs = append(errs, err)
	}
	labelExclusion, err := labelExclusion(configuration)
	if err != nil {
		errs = append(errs, err)
	}
	rule, err := getRules(configuration)
	if err != nil {
		errs = append(errs, err)
	}
	return options{
		namespace:      namespace,
		pollInterval:   pollInterval,
		runDuration:    runDuration,
		labelExclusion: labelExclusion,
		rules:          rule,
	}
}
