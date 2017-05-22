package main

import (
	"os"
	"testing"
	"time"
	"github.com/target/pod-reaper/rules"
)

func TestDefaultNamespace(test *testing.T) {
	os.Clearenv()
	namespace := namespace()
	if namespace != "" {
		test.Errorf("EXPECTED: \"\" ACTUAL: \"%s\"", namespace)
	}
}

func TestNamespace(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_NAMESPACE, "test-namespace")
	namespace := namespace()
	if namespace != "test-namespace" {
		test.Errorf("EXPECTED: \"test-namespace\" ACTUAL: \"%s\"", namespace)
	}
}

func TestDefaultPollInterval(test *testing.T) {
	os.Clearenv()
	duration, err := pollInterval()
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
	if duration != time.Minute {
		test.Errorf("EXPECTED: \"1m\" ACTUAL: \"%s\"", duration.String())
	}
}

func TestInvalidPollInterval(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_POLL_INTERVAL, "not-a-duration")
	_, err := pollInterval()
	if err == nil {
		test.Error("expected error")
	}
}

func TestPollInterval(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_POLL_INTERVAL, "1m58s")
	duration, err := pollInterval()
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
	if duration != (2 * (time.Minute - time.Second)) {
		test.Errorf("EXPECTED: \"1m58s\" ACTUAL: \"%s\"", duration.String())
	}
}

func TestDefaultRunDuration(test *testing.T) {
	os.Clearenv()
	duration, err := runDuration()
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
	if duration != 0 {
		test.Errorf("EXPECTED: \"0s\" ACTUAL: \"%s\"", duration.String())
	}
}

func TestInvalidRunDuration(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_RUN_DURATION, "not-a-duration")
	_, err := runDuration()
	if err == nil {
		test.Error("expected error")
	}
}

func TestRunDuration(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_RUN_DURATION, "1m58s")
	duration, err := runDuration()
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
	if duration != (2 * (time.Minute - time.Second)) {
		test.Errorf("EXPECTED: \"1m58s\" ACTUAL: \"%s\"", duration.String())
	}
}

func TestDefaultLabelExclusion(test *testing.T) {
	os.Clearenv()
	exclusion, err := labelExclusion()
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
	if exclusion != nil {
		test.Error("expected nil")
	}
}

func TestKeyOnlyLabelExclusion(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_EXCLUDE_LABEL_KEY, "test-key")
	_, err := labelExclusion()
	if err == nil {
		test.Error("expected error")
	}
}

func TestValuesOnlyLabelExclusion(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_EXCLUDE_LABEL_VALUES, "test-value1,test-value2")
	_, err := labelExclusion()
	if err == nil {
		test.Error("expected error")
	}
}

func TestInvalidLabelExclusion(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_EXCLUDE_LABEL_KEY, "keys cannot have spaces")
	os.Setenv(ENV_EXCLUDE_LABEL_VALUES, "test-value1,test-value2")
	_, err := labelExclusion()
	if err == nil {
		test.Error("expected error")
	}
}

func TestLabelExclusion(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_EXCLUDE_LABEL_KEY, "test-key")
	os.Setenv(ENV_EXCLUDE_LABEL_VALUES, "test-value1,test-value2")
	exclusion, err := labelExclusion()
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
	if exclusion == nil {
		test.Error("expected not nil")
	}
}

func TestOptionsLoadInvalidPollInterval(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_POLL_INTERVAL, "invalid")
	_, err := loadOptions()
	if err == nil {
		test.Error("expected error")
	}
}
func TestOptionsLoadInvalidDuration(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_RUN_DURATION, "invalid")
	_, err := loadOptions()
	if err == nil {
		test.Error("expected error")
	}
}
func TestOptionsLoadInvalidLabelExclusion(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_EXCLUDE_LABEL_KEY, "not a valid key")
	_, err := loadOptions()
	if err == nil {
		test.Error("expected error")
	}
}

func TestOptionsLoadInvalidRules(test *testing.T) {
	os.Clearenv()
	// no rules defined
	_, err := loadOptions()
	if err == nil {
		test.Error("expected error")
	}
}

func TestOptionsLoad(test *testing.T) {
	os.Clearenv()
	// ensure at least one rule loads
	os.Setenv(rules.ENV_CHAOS_CHANCE, "1.0")
	options, err := loadOptions()
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
	if options.pollInterval != time.Minute {
		test.Errorf("EXPECTED \"1m\" ACTUAL: \"%s\"", options.pollInterval.String())
	}
	if options.runDuration != 0 {
		test.Errorf("EXPECTED \"0s\" ACTUAL: \"%s\"", options.pollInterval.String())
	}
	if options.labelExclusion != nil {
		test.Error("expected nil")
	}
	if len(options.rules.LoadedRules) < 1 {
		test.Error("expected at least one rule to be loaded")
	}
}
