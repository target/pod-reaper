package rules

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

func testDurationPod(startTime *time.Time) v1.Pod {
	pod := v1.Pod{}
	if startTime != nil {
		setTime := unversioned.NewTime(*startTime)
		pod.Status.StartTime = &setTime
	}
	return pod
}

func TestDurationLoad(t *testing.T) {
	t.Run("load", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(ENV_MAX_DURATION, "30m")
		loaded, err := (&duration{}).load()
		assert.NoError(t, err)
		assert.True(t, loaded)
	})
	t.Run("invalid duration", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(ENV_MAX_DURATION, "not-a-duration")
		loaded, err := (&duration{}).load()
		assert.Error(t, err)
		assert.False(t, loaded)
	})
	t.Run("no load", func(t *testing.T) {
		os.Clearenv()
		loaded, err := (&duration{}).load()
		assert.NoError(t, err)
		assert.False(t, loaded)
	})
}

func TestDurationShouldReap(t *testing.T) {
	t.Run("no start time", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(ENV_MAX_DURATION, "2m")
		duration := duration{}
		duration.load()
		pod := testDurationPod(nil) // no start time can happen during pod creation
		shouldReap, _ := duration.ShouldReap(pod)
		assert.False(t, shouldReap)
	})
	t.Run("reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(ENV_MAX_DURATION, "1m59s")
		duration := duration{}
		duration.load()
		startTime := time.Now().Add(-2 * time.Minute)
		pod := testDurationPod(&startTime)
		shouldReap, reason := duration.ShouldReap(pod)
		assert.True(t, shouldReap)
		assert.Regexp(t, ".*has been running.*", reason)
	})
	t.Run("no reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(ENV_MAX_DURATION, "2m1s")
		duration := duration{}
		duration.load()
		startTime := time.Now().Add(-2 * time.Minute)
		pod := testDurationPod(&startTime)
		shouldReap, _ := duration.ShouldReap(pod)
		assert.False(t, shouldReap)
	})
}
