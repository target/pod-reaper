package rules

import (
	"errors"

	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

// Rule is an interface defining the two functions needed for pod reaper to use the rule.
type Rule interface {

	// load attempts to load the load and returns whether the rule was loaded, a message that will be logged
	// when the rule is loaded, and any error that may have occurred during the load.
	load() (bool, string, error)

	// ShouldReap takes a pod and returns whether the pod should be reaped based on this rule and a message that
	// will be logged when the pod is selected for reaping.
	ShouldReap(pod v1.Pod) (bool, string)
}

// Rules is a collection of loaded pod reaper rules.
type Rules struct {
	LoadedRules []Rule
}

// LoadRules load all the rules based on their own implementations
func LoadRules() (Rules, error) {
	// load all possible rules
	rules := []Rule{
		&chaos{},
		&containerStatus{},
		&duration{},
		&unready{},
		&podStatus{},
	}
	// return only the active rules
	loadedRules := []Rule{}
	for _, rule := range rules {
		load, message, err := rule.load()
		if err != nil {
			return Rules{LoadedRules: loadedRules}, err
		} else if load {
			logrus.Info("loaded rule: " + message)
			loadedRules = append(loadedRules, rule)
		}
	}
	// return an error if no rules where loaded
	if len(loadedRules) == 0 {
		return Rules{LoadedRules: loadedRules}, errors.New("no rules were loaded")
	}
	return Rules{LoadedRules: loadedRules}, nil
}

// ShouldReap takes a pod and return whether the pod should be reaped based on this rule.
// Also includes a message describing why the pod was flagged for reaping.
func (rules Rules) ShouldReap(pod v1.Pod) (bool, []string) {
	var reasons []string
	for _, rule := range rules.LoadedRules {
		reap, reason := rule.ShouldReap(pod)
		if !reap {
			return false, []string{}
		}
		reasons = append(reasons, reason)
	}
	return true, reasons
}
