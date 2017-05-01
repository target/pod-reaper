package rules

import (
	"k8s.io/client-go/pkg/api/v1"
	"errors"
)

type Rule interface {
	load() (bool, error)
	ShouldReap(pod v1.Pod) (bool, string)
}

func LoadRules() ([]Rule, error) {
	// load all possible rules
	rules := []Rule{
		&duration{},
		&containerStatus{},
		&chaos{},
	}
	// return only the active rules
	loadedRules := []Rule{}
	for _, rule := range rules {
		load, err := rule.load()
		if err != nil {
			return loadedRules, err
		} else if load {
			loadedRules = append(loadedRules, rule)
		}
	}
	// return an err if no rules where loaded
	if len(loadedRules) == 0 {
		return loadedRules, errors.New("no rules were loaded")
	}
	return loadedRules, nil
}
