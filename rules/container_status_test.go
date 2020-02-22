package rules

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/client-go/pkg/api/v1"
)

func TestContainerStatusIgnore(t *testing.T) {
	os.Unsetenv(envContainerStatus)
	reapResult, message := containerStatus(v1.Pod{})
	assert.Equal(t, ignore, reapResult)
	assert.Equal(t, notConfigured, message)
}

func TestContainerStatus(t *testing.T) {
	tests := []struct {
		env               string
		containerStatuses []v1.ContainerState
		reapResult        result
		message           string
	}{
		{
			env:               "test",
			containerStatuses: []v1.ContainerState{waiting("other")},
			reapResult:        spare,
			message:           "has no container with status in {test}",
		},
		{
			env:               "test",
			containerStatuses: []v1.ContainerState{terminated("other")},
			reapResult:        spare,
			message:           "has no container with status in {test}",
		},
		{
			env:               "test",
			containerStatuses: []v1.ContainerState{waiting("test")},
			reapResult:        reap,
			message:           "has container with status 'test' in {test}",
		},
		{
			env:               "test",
			containerStatuses: []v1.ContainerState{terminated("test")},
			reapResult:        reap,
			message:           "has container with status 'test' in {test}",
		},
		{
			env:               "test,second",
			containerStatuses: []v1.ContainerState{terminated("second")},
			reapResult:        reap,
			message:           "has container with status 'second' in {test,second}",
		},
		{
			env:               "test,second",
			containerStatuses: []v1.ContainerState{terminated("other")},
			reapResult:        spare,
			message:           "has no container with status in {test,second}",
		},
	}
	for _, test := range tests {
		os.Setenv(envContainerStatus, test.env)
		pod := containerStatePod(test.containerStatuses)
		reapResult, message := containerStatus(pod)
		assert.Equal(t, test.reapResult, reapResult)
		assert.Equal(t, test.message, message)
	}
}

func waiting(reason string) v1.ContainerState {
	return v1.ContainerState{
		Waiting: &v1.ContainerStateWaiting{
			Reason: reason,
		},
	}
}

func terminated(reason string) v1.ContainerState {
	return v1.ContainerState{
		Terminated: &v1.ContainerStateTerminated{
			Reason: reason,
		},
	}
}

func containerStatePod(containerStates []v1.ContainerState) v1.Pod {
	var statuses = []v1.ContainerStatus{}
	for _, state := range containerStates {
		statuses = append(statuses, v1.ContainerStatus{State: state})
	}
	return v1.Pod{
		Status: v1.PodStatus{
			ContainerStatuses: statuses,
		},
	}
}
