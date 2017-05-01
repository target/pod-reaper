package rules

import (
	"testing"
	"os"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/api/unversioned"
	"time"
	"strings"
)

func testPod(startTime *time.Time) v1.Pod {
	pod := v1.Pod{}
	if startTime != nil {
		setTime := unversioned.NewTime(*startTime)
		pod.Status.StartTime = &setTime
	}
	return pod
}

func TestLoad(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "30m")
	active, err := (&duration{}).load()
	if !active || err != nil {
		test.Fail()
	}
}

func TestFailLoad(test *testing.T) {
	os.Clearenv()
	active, err := (&duration{}).load()
	if active || err != nil {
		test.Fail()
	}
}

func TestInvalid(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "not-a-duration")
	active, err := (&duration{}).load()
	if active || err == nil {
		test.Fail()
	}
}

func TestNoStartTime(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "2m")
	duration := duration{}
	duration.load()
	pod := testPod(nil) // no start time
	shouldReap, _ := duration.ShouldReap(pod)
	if shouldReap {
		test.Fail()
	}
}

func TestShouldReap(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "1m59s")
	duration := duration{}
	duration.load()
	startTime := time.Now().Add(-2 * time.Minute)
	pod := testPod(&startTime)
	shouldReap, message := duration.ShouldReap(pod)
	if !shouldReap || !strings.Contains(message, "has been running") {
		test.Fail()
	}
}

func TestShouldNotReap(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_MAX_DURATION, "2m1s")
	duration := duration{}
	duration.load()
	startTime := time.Now().Add(-2 * time.Minute)
	pod := testPod(&startTime)
	shouldReap, _ := duration.ShouldReap(pod)
	if shouldReap {
		test.Fail()
	}
}
