package rules

import (
	"k8s.io/client-go/pkg/api/v1"
	"errors"
)

type Rule interface {
	load() bool
	ShouldReap(pod v1.Pod) (bool, string)
}

func LoadRules() []Rule {
	// load all possible rules
	rules := []Rule{
		&maxDurationRule{},
		&containerStatusesRule{},
		&chaosRule{},
	}
	// return only the active rules
	loadedRules := []Rule{}
	for _, rule := range rules {
		if rule.load() {
			loadedRules = append(loadedRules, rule)
		}
	}
	//  panic if no rules are loaded
	if len(loadedRules) == 0 {
		panic(errors.New("no rules were loaded"))
	}
	return loadedRules
}
