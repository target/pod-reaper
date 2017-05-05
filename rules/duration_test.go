package rules

import (
	"testing"
	"os"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/api/unversioned"
	"time"
	"strings"
)

func testDurationPod(startTime *time.Time) v1.Pod {
	pod := v1.Pod{}
	if startTime != nil {
		setTime := unversioned.NewTime(*startTime)
		pod.Status.StartTime = &setTime
	}
	return pod
}

func TestDurationLoad(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "30m")
	loaded, err := (&duration{}).load()
	if !loaded {
		test.Error("not loaded")
	}
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
}

func TestDurationFailLoad(test *testing.T) {
	os.Clearenv()
	loaded, err := (&duration{}).load()
	if loaded {
		test.Error("loaded")
	}
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
}

func TestDurationInvalid(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "not-a-duration")
	loaded, err := (&duration{}).load()
	if loaded {
		test.Error("loaded")
	}
	if err == nil {
		test.Error("expected error")
	}
}

func TestDurationNoStartTime(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "2m")
	duration := duration{}
	duration.load()
	pod := testDurationPod(nil) // no start time
	shouldReap, _ := duration.ShouldReap(pod)
	if shouldReap {
		test.Error("should reap")
	}
}

func TestDurationShouldReap(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "1m59s")
	duration := duration{}
	duration.load()
	startTime := time.Now().Add(-2 * time.Minute)
	pod := testDurationPod(&startTime)
	shouldReap, message := duration.ShouldReap(pod)
	if !shouldReap {
		test.Error("should not reap")
	}
	if !strings.Contains(message, "has been running") {
		test.Errorf("EXPECTED \"has been running\" CONTAINED IN: %s", message)
	}
}

func TestDurationShouldNotReap(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "2m1s")
	duration := duration{}
	duration.load()
	startTime := time.Now().Add(-2 * time.Minute)
	pod := testDurationPod(&startTime)
	shouldReap, _ := duration.ShouldReap(pod)
	if shouldReap {
		test.Error("should reap")
	}
}
