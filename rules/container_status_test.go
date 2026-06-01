package rules

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func testWaitContainerState(reason string) v1.ContainerState {
	return v1.ContainerState{
		Waiting: &v1.ContainerStateWaiting{
			Reason: reason,
		},
	}
}

func testTerminatedContainerState(reason string) v1.ContainerState {
	return v1.ContainerState{
		Terminated: &v1.ContainerStateTerminated{
			Reason: reason,
		},
	}

}

func testStatusPod(containerState v1.ContainerState) v1.Pod {
	return v1.Pod{
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: containerState,
				},
			},
		},
	}
}

func TestContainerStatusLoad(t *testing.T) {
	t.Run("load", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envContainerStatus, "test-status")
		loaded, message, err := (&containerStatus{}).load()
		assert.NoError(t, err)
		assert.Equal(t, "container status in [test-status]", message)
		assert.True(t, loaded)
	})
	t.Run("load multiple-statuses", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envContainerStatus, "test-status,another-status")
		containerStatus := containerStatus{}
		loaded, message, err := containerStatus.load()
		assert.NoError(t, err)
		assert.Equal(t, "container status in [test-status,another-status]", message)
		assert.True(t, loaded)
		assert.Equal(t, 2, len(containerStatus.reapStatuses))
		assert.Equal(t, "test-status", containerStatus.reapStatuses[0])
		assert.Equal(t, "another-status", containerStatus.reapStatuses[1])
	})
	t.Run("no load", func(t *testing.T) {
		os.Clearenv()
		loaded, message, err := (&containerStatus{}).load()
		assert.NoError(t, err)
		assert.Equal(t, "", message)
		assert.False(t, loaded)
	})
}

func TestContainerStatusShouldReap(t *testing.T) {
	t.Run("reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envContainerStatus, "test-status,another-status")
		containerStatus := containerStatus{}
		containerStatus.load()
		pod := testStatusPod(testWaitContainerState("another-status"))
		shouldReap, reason := containerStatus.ShouldReap(pod)
		assert.True(t, shouldReap)
		assert.Regexp(t, ".*another-status.*", reason)
	})
	t.Run("no reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envContainerStatus, "test-status,another-status")
		containerStatus := containerStatus{}
		containerStatus.load()
		pod := testStatusPod(testWaitContainerState("not-present"))
		shouldReap, _ := containerStatus.ShouldReap(pod)
		assert.False(t, shouldReap)
	})
	t.Run("reap terminated state", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envContainerStatus, "Error")
		cs := containerStatus{}
		cs.load()
		pod := testStatusPod(testTerminatedContainerState("Error"))
		shouldReap, reason := cs.ShouldReap(pod)
		assert.True(t, shouldReap)
		assert.Regexp(t, ".*Error.*", reason)
	})
	t.Run("init container only - matches", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envContainerStatus, "CrashLoopBackOff")
		cs := containerStatus{}
		cs.load()
		pod := v1.Pod{
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{}, // no regular containers
				InitContainerStatuses: []v1.ContainerStatus{
					{
						State: testWaitContainerState("CrashLoopBackOff"),
					},
				},
			},
		}
		shouldReap, reason := cs.ShouldReap(pod)
		assert.True(t, shouldReap)
		assert.Contains(t, reason, "init container")
	})
	t.Run("no containers returns false", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envContainerStatus, "CrashLoopBackOff")
		cs := containerStatus{}
		cs.load()
		pod := v1.Pod{
			Status: v1.PodStatus{
				ContainerStatuses:     []v1.ContainerStatus{},
				InitContainerStatuses: []v1.ContainerStatus{},
			},
		}
		shouldReap, _ := cs.ShouldReap(pod)
		assert.False(t, shouldReap)
	})
	t.Run("running state does not match", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envContainerStatus, "Running")
		cs := containerStatus{}
		cs.load()
		pod := v1.Pod{
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
			},
		}
		shouldReap, _ := cs.ShouldReap(pod)
		assert.False(t, shouldReap) // Running state has no Reason field to match
	})
	t.Run("whitespace in values not trimmed", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envContainerStatus, "Status1, Status2")
		cs := containerStatus{}
		cs.load()
		// The second value is " Status2" with leading space
		assert.Equal(t, " Status2", cs.reapStatuses[1])
		// Pod with "Status2" (no space) won't match " Status2"
		pod := testStatusPod(testWaitContainerState("Status2"))
		shouldReap, _ := cs.ShouldReap(pod)
		assert.False(t, shouldReap)
	})
}
