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
}
