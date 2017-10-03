package rules

import (
	"errors"
	"k8s.io/client-go/pkg/api/v1"
	"strings"
)

// Rule is an interface defining the two functions needed for pod reaper to use the rule.
type Rule interface {
	load() (bool, error)
	ShouldReap(pod v1.Pod) (bool, string)
}

// Rules is a collection of loaded pod reaper rules.
type Rules struct {
	LoadedRules []Rule
}

// LoadRules load all of the rules based on their own implementations
func LoadRules() (Rules, error) {
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
			return Rules{LoadedRules: loadedRules}, err
		} else if load {
			loadedRules = append(loadedRules, rule)
		}
	}
	// return an err if no rules where loaded
	if len(loadedRules) == 0 {
		return Rules{LoadedRules: loadedRules}, errors.New("no rules were loaded")
	}
	return Rules{LoadedRules: loadedRules}, nil
}

// ShouldReap takes a pod and return whether or not the pod should be reaped based on this rule.
// Also includes a message describing why the pod was flagged for reaping.
func (rules Rules) ShouldReap(pod v1.Pod) (bool, string) {
	reasons := []string{}
	for _, rule := range rules.LoadedRules {
		reap, reason := rule.ShouldReap(pod)
		if !reap {
			return false, ""
		}
		reasons = append(reasons, reason)
	}
	return true, strings.Join(reasons, " AND ")
}
