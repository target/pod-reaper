package rules

import (
	"os"
	"testing"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/api/unversioned"
	"time"
	"strings"
)

func TestLoadNoRules(test *testing.T) {
	os.Clearenv()
	rules, err := LoadRules()
	if len(rules.LoadedRules) != 0 {
		test.Errorf("rules were loaded: %s", rules)
	}
	if err == nil {
		test.Error("expected error")
	}
}

func TestLoadInvalidRules(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CHAOS_CHANCE, "not-a-number")
	_, err := LoadRules()
	if err == nil {
		test.Error("expected error")
	}
}

func TestLoadRules(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "2m")
	os.Setenv(ENV_CONTAINER_STATUS, "test-status")
	rules, err := LoadRules()
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
	if len(rules.LoadedRules) != 2 {
		test.Errorf("EXPECTED: 2 ACTUAL: %d", len(rules.LoadedRules))
	}
}

func testPod() v1.Pod {
	startTime := unversioned.NewTime(time.Now().Add(-2 * time.Minute))
	return v1.Pod{
		Status:v1.PodStatus{
			StartTime: &startTime,
			ContainerStatuses:[]v1.ContainerStatus{
				{
					State: testTerminatedContainerState("test-status"),
				},
			},
		},
	}
}

func TestShouldReap(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "1m59s")
	os.Setenv(ENV_CONTAINER_STATUS, "test-status")
	os.Setenv(ENV_CHAOS_CHANCE, "1.0") // always
	loaded, _ := LoadRules()
	shouldReap, message := loaded.ShouldReap(testPod())
	if !shouldReap {
		test.Error("should not reap")
	}
	if !strings.Contains(message, "was flagged for chaos") {
		test.Errorf("EXPECTED \"was flagged for chaos\" CONTAINED IN %s", message)
	}
	if !strings.Contains(message, "has status test-status") {
		test.Errorf("EXPECTED \"has status test-status\" CONTAINED IN %s", message)
	}
	if !strings.Contains(message, "has been running") {
		test.Errorf("EXPECTED \"has been running\" CONTAINED IN %s", message)
	}
}

func TestShouldNotReap(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "1m59s")
	os.Setenv(ENV_CONTAINER_STATUS, "test-status")
	os.Setenv(ENV_CHAOS_CHANCE, "0.0") // never
	loaded, _ := LoadRules()
	shouldReap, _ := loaded.ShouldReap(testPod())
	if shouldReap {
		test.Error("should reap")
	}
}
