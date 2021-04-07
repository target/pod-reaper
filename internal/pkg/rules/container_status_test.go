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
}
