package main

import (
	"testing"
	"time"
)

func TestDefaultMaxPodDuration(test *testing.T) {
	duration, err := time.ParseDuration("2h")
	if err != nil {
		test.Errorf(err.Error())
	}
	if options().maxPodDuration.String() != duration.String() {
		test.Fail()
	}
}

func TestDefaultPollInterval(test *testing.T) {
	duration, err := time.ParseDuration("1m")
	if err != nil {
		test.Errorf(err.Error())
	}
	if options().pollInterval.String() != duration.String() {
		test.Fail()
	}
}

func TestDefaultContainerState(test *testing.T) {
	if len(options().containerStatuses) > 0 {
		test.Fail()
	}
}

func TestDefaultExcludeLabelKey(test *testing.T) {
	if options().excludeLabelKey != "pod-reaper" {
		test.Fail()
	}
}

func TestDefaultExcludeLabelValue(test *testing.T) {
	if options().excludeLabelValue != "disabled" {
		test.Fail()
	}
}

func TestDefaultNamespace(test *testing.T) {
	if options().namespace != "" {
		test.Fail()
	}
}
