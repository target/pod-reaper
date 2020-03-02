package rules

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	k8v1 "k8s.io/client-go/pkg/api/v1"
)

func TestContainerStatusIgnore(t *testing.T) {
	os.Unsetenv(envContainerStatus)
	reapResult, message := containerStatus(k8v1.Pod{})
	assert.Equal(t, ignore, reapResult)
	assert.Equal(t, notConfigured, message)
}

func TestContainerStatus(t *testing.T) {
	tests := []struct {
		env               string
		containerStatuses []k8v1.ContainerState
		reapResult        result
		message           string
	}{
		{
			env:               "test",
			containerStatuses: []k8v1.ContainerState{waiting("other")},
			reapResult:        spare,
			message:           "has no container with status in {test}",
		},
		{
			env:               "test",
			containerStatuses: []k8v1.ContainerState{terminated("other")},
			reapResult:        spare,
			message:           "has no container with status in {test}",
		},
		{
			env:               "test",
			containerStatuses: []k8v1.ContainerState{waiting("test")},
			reapResult:        reap,
			message:           "has container with status 'test' in {test}",
		},
		{
			env:               "test",
			containerStatuses: []k8v1.ContainerState{terminated("test")},
			reapResult:        reap,
			message:           "has container with status 'test' in {test}",
		},
		{
			env:               "test,second",
			containerStatuses: []k8v1.ContainerState{terminated("second")},
			reapResult:        reap,
			message:           "has container with status 'second' in {test,second}",
		},
		{
			env:               "test,second",
			containerStatuses: []k8v1.ContainerState{terminated("other")},
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

func waiting(reason string) k8v1.ContainerState {
	return k8v1.ContainerState{
		Waiting: &k8v1.ContainerStateWaiting{
			Reason: reason,
		},
	}
}

func terminated(reason string) k8v1.ContainerState {
	return k8v1.ContainerState{
		Terminated: &k8v1.ContainerStateTerminated{
			Reason: reason,
		},
	}
}

func containerStatePod(containerStates []k8v1.ContainerState) k8v1.Pod {
	var statuses = []k8v1.ContainerStatus{}
	for _, state := range containerStates {
		statuses = append(statuses, k8v1.ContainerStatus{State: state})
	}
	return k8v1.Pod{
		Status: k8v1.PodStatus{
			ContainerStatuses: statuses,
		},
	}
}
