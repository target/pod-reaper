package rules

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

func testPod() v1.Pod {
	startTime := unversioned.NewTime(time.Now().Add(-2 * time.Minute))
	return v1.Pod{
		Status: v1.PodStatus{
			StartTime: &startTime,
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: testTerminatedContainerState("test-status"),
				},
			},
		},
	}
}

func TestRules(t *testing.T) {
	t.Run("no rules", func(t *testing.T) {
		os.Clearenv()
		rules, err := LoadRules()
		assert.Equal(t, 0, len(rules.LoadedRules))
		assert.Error(t, err)
	})
	t.Run("invalid rule", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envChaosChance, "not-a-number")
		_, err := LoadRules()
		assert.Error(t, err)
	})
	t.Run("load", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envMaxDuration, "2m")
		os.Setenv(envContainerStatus, "test-status")
		rules, err := LoadRules()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(rules.LoadedRules))
	})
}

func TestShouldReap(t *testing.T) {
	t.Run("reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envMaxDuration, "1m59s")
		os.Setenv(envContainerStatus, "test-status")
		os.Setenv(envChaosChance, "1.0") // always
		loaded, _ := LoadRules()
		shouldReap, message := loaded.ShouldReap(testPod())
		assert.True(t, shouldReap)
		assert.Regexp(t, ".*was flagged for chaos.*", message)
		assert.Regexp(t, ".*has status test-status.*", message)
		assert.Regexp(t, ".*has been running.*", message)
	})
	t.Run("no reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envMaxDuration, "1m59s")
		os.Setenv(envContainerStatus, "test-status")
		os.Setenv(envChaosChance, "0.0") // never
		loaded, _ := LoadRules()
		shouldReap, _ := loaded.ShouldReap(testPod())
		assert.False(t, shouldReap)
	})
}
