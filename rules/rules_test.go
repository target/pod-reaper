package rules

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	logrus.SetOutput(ioutil.Discard) // disable logging during tests
}

func testPod() v1.Pod {
	startTime := metav1.NewTime(time.Now().Add(-2 * time.Minute))
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
		shouldReap, reasons := loaded.ShouldReap(testPod())
		assert.True(t, shouldReap)
		if assert.Equal(t, 3, len(reasons)) {
			assert.Regexp(t, ".*was flagged for chaos.*", reasons[0])
			assert.Regexp(t, ".*test-status.*", reasons[1])
			assert.Regexp(t, ".*has been running.*", reasons[2])
		}
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
	t.Run("all 6 rules loaded", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envChaosChance, "1.0")
		os.Setenv(envContainerStatus, "CrashLoopBackOff")
		os.Setenv(envMaxDuration, "1h")
		os.Setenv(envMaxUnready, "10m")
		os.Setenv(envPodStatus, "Evicted")
		os.Setenv(envPodStatusPhase, "Failed")
		rules, err := LoadRules()
		assert.NoError(t, err)
		assert.Equal(t, 6, len(rules.LoadedRules))
	})
	t.Run("AND logic - one false rule prevents reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envChaosChance, "0.0")       // always false
		os.Setenv(envContainerStatus, "Error") // would match
		os.Setenv(envMaxDuration, "1s")        // would match (pod is 2min old)
		rules, _ := LoadRules()
		shouldReap, reasons := rules.ShouldReap(testPod())
		assert.False(t, shouldReap)
		assert.Empty(t, reasons) // no reasons when short-circuited
	})
	t.Run("complex pod with init containers and multiple conditions", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envContainerStatus, "ImagePullBackOff")
		os.Setenv(envMaxDuration, "1m")
		rules, err := LoadRules()
		assert.NoError(t, err)

		// Create complex pod
		startTime := metav1.NewTime(time.Now().Add(-5 * time.Minute))
		transitionTime := metav1.NewTime(time.Now().Add(-3 * time.Minute))
		pod := v1.Pod{
			Status: v1.PodStatus{
				StartTime: &startTime,
				Phase:     v1.PodPending,
				Reason:    "ImagePullBackOff",
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "main",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
				InitContainerStatuses: []v1.ContainerStatus{
					{
						Name:  "init",
						State: testWaitContainerState("ImagePullBackOff"),
					},
				},
				Conditions: []v1.PodCondition{
					{
						Type:               v1.PodReady,
						Status:             "False",
						LastTransitionTime: transitionTime,
					},
					{
						Type:   v1.PodInitialized,
						Status: "False",
					},
				},
			},
		}
		shouldReap, reasons := rules.ShouldReap(pod)
		assert.True(t, shouldReap)
		assert.Equal(t, 2, len(reasons))
		// Check that init container status was matched
		assert.Contains(t, reasons[0], "init container")
	})
}
