package rules

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func testPodFromReason(reason string) v1.Pod {
	return v1.Pod{
		Status: v1.PodStatus{
			Reason: reason,
		},
	}
}

func TestPodStatusLoad(t *testing.T) {
	t.Run("load", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envPodStatus, "test-status")
		loaded, message, err := (&podStatus{}).load()
		assert.NoError(t, err)
		assert.Equal(t, "pod status in [test-status]", message)
		assert.True(t, loaded)
	})
	t.Run("load multiple-statuses", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envPodStatus, "test-status,another-status")
		podStatus := podStatus{}
		loaded, message, err := podStatus.load()
		assert.NoError(t, err)
		assert.Equal(t, "pod status in [test-status,another-status]", message)
		assert.True(t, loaded)
		assert.Equal(t, 2, len(podStatus.reapStatuses))
		assert.Equal(t, "test-status", podStatus.reapStatuses[0])
		assert.Equal(t, "another-status", podStatus.reapStatuses[1])
	})
	t.Run("no load", func(t *testing.T) {
		os.Clearenv()
		loaded, message, err := (&podStatus{}).load()
		assert.NoError(t, err)
		assert.Equal(t, "", message)
		assert.False(t, loaded)
	})
}

func TestPodStatusShouldReap(t *testing.T) {
	t.Run("reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envPodStatus, "test-status,another-status")
		podStatus := podStatus{}
		podStatus.load()
		pod := testPodFromReason("another-status")
		shouldReap, reason := podStatus.ShouldReap(pod)
		assert.True(t, shouldReap)
		assert.Regexp(t, ".*another-status.*", reason)
	})
	t.Run("no reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envPodStatus, "test-status,another-status")
		podStatus := podStatus{}
		podStatus.load()
		pod := testPodFromReason("not-present")
		shouldReap, _ := podStatus.ShouldReap(pod)
		assert.False(t, shouldReap)
	})
}
