package rules

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func testPodFromPhase(phase v1.PodPhase) v1.Pod {
	return v1.Pod{
		Status: v1.PodStatus{
			Phase: phase,
		},
	}
}

func TestPodPhaseLoad(t *testing.T) {
	t.Run("load", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envPodStatusPhase, "test-phase")
		loaded, message, err := (&podStatusPhase{}).load()
		assert.NoError(t, err)
		assert.Equal(t, "pod status phase in [test-phase]", message)
		assert.True(t, loaded)
	})
	t.Run("load multiple-statuses", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envPodStatusPhase, "test-phase,another-phase")
		podStatusPhase := podStatusPhase{}
		loaded, message, err := podStatusPhase.load()
		assert.NoError(t, err)
		assert.Equal(t, "pod status phase in [test-phase,another-phase]", message)
		assert.True(t, loaded)
		assert.Equal(t, 2, len(podStatusPhase.reapStatusPhases))
		assert.Equal(t, "test-phase", podStatusPhase.reapStatusPhases[0])
		assert.Equal(t, "another-phase", podStatusPhase.reapStatusPhases[1])
	})
	t.Run("no load", func(t *testing.T) {
		os.Clearenv()
		loaded, message, err := (&podStatusPhase{}).load()
		assert.NoError(t, err)
		assert.Equal(t, "", message)
		assert.False(t, loaded)
	})
}

func TestPodStatusPhaseShouldReap(t *testing.T) {
	t.Run("reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envPodStatusPhase, "test-phase,another-phase")
		podStatusPhase := podStatusPhase{}
		podStatusPhase.load()
		pod := testPodFromPhase("another-phase")
		shouldReap, reason := podStatusPhase.ShouldReap(pod)
		assert.True(t, shouldReap)
		assert.Regexp(t, ".*another-phase.*", reason)
	})
	t.Run("no reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envPodStatusPhase, "test-phase,another-phase")
		podStatusPhase := podStatusPhase{}
		podStatusPhase.load()
		pod := testPodFromPhase("not-present")
		shouldReap, _ := podStatusPhase.ShouldReap(pod)
		assert.False(t, shouldReap)
	})
	t.Run("whitespace in values not trimmed", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envPodStatusPhase, "Failed, Unknown")
		psp := podStatusPhase{}
		psp.load()
		// The second value is " Unknown" with leading space
		assert.Equal(t, " Unknown", psp.reapStatusPhases[1])
		// Pod with "Unknown" (no space) won't match " Unknown"
		pod := testPodFromPhase(v1.PodUnknown)
		shouldReap, _ := psp.ShouldReap(pod)
		assert.False(t, shouldReap)
	})
	t.Run("case sensitivity - lowercase fails to match", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envPodStatusPhase, "failed")
		psp := podStatusPhase{}
		psp.load()
		pod := testPodFromPhase(v1.PodFailed) // "Failed" in K8s
		shouldReap, _ := psp.ShouldReap(pod)
		assert.False(t, shouldReap) // "failed" != "Failed"
	})
	t.Run("all valid phases", func(t *testing.T) {
		phases := []v1.PodPhase{
			v1.PodPending,
			v1.PodRunning,
			v1.PodSucceeded,
			v1.PodFailed,
			v1.PodUnknown,
		}
		for _, phase := range phases {
			t.Run(string(phase), func(t *testing.T) {
				os.Clearenv()
				os.Setenv(envPodStatusPhase, string(phase))
				psp := podStatusPhase{}
				psp.load()
				pod := testPodFromPhase(phase)
				shouldReap, reason := psp.ShouldReap(pod)
				assert.True(t, shouldReap)
				assert.Contains(t, reason, string(phase))
			})
		}
	})
}
