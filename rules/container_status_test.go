package rules

import (
	"k8s.io/client-go/pkg/api/v1"
	"os"
	"testing"
	"strings"
)

func testWaitContainerState(reason string) v1.ContainerState {
	return v1.ContainerState{
		Waiting:&v1.ContainerStateWaiting{
			Reason: reason,
		},
	}
}

func testTerminatedContainerState(reason string) v1.ContainerState {
	return v1.ContainerState{
		Terminated:&v1.ContainerStateTerminated{
			Reason: reason,
		},
	}

}

func testStatusPod(containerState v1.ContainerState) v1.Pod {
	return v1.Pod{
		Status:v1.PodStatus{
			ContainerStatuses:[]v1.ContainerStatus{
				{
					State: containerState,
				},
			},
		},
	}
}

func TestStatusLoad(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CONTAINER_STATUS, "test-status")
	active, err := (&containerStatus{}).load()
	if !active || err != nil {
		test.Fail()
	}
}

func TestStatusLoadMultiple(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CONTAINER_STATUS, "test-status,another-status")
	containerStatus := containerStatus{}
	containerStatus.load()
	statuses := containerStatus.reapStatuses
	if len(statuses) != 2 || statuses[0] != "test-status" || statuses[1] != "another-status" {
		test.Fail()
	}
}

func TestStatusFailLoad(test *testing.T) {
	os.Clearenv()
	active, err := (&containerStatus{}).load()
	if active || err != nil {
		test.Fail()
	}
}



func TestStatusWaiting(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CONTAINER_STATUS, "test-status,another-status")
	containerStatus := containerStatus{}
	containerStatus.load()
	pod := testStatusPod(testWaitContainerState("test-status"))
	shouldReap, reason := containerStatus.ShouldReap(pod)
	if !shouldReap || !strings.Contains(reason, "test-status") {
		test.Fail()
	}
}

func TestStatusTerminated(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CONTAINER_STATUS, "test-status,another-status")
	containerStatus := containerStatus{}
	containerStatus.load()
	pod := testStatusPod(testTerminatedContainerState("another-status"))
	shouldReap, reason := containerStatus.ShouldReap(pod)
	if !shouldReap || !strings.Contains(reason, "another-status") {
		test.Fail()
	}
}

func TestStatusNotPresent(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CONTAINER_STATUS, "test-status,another-status")
	containerStatus := containerStatus{}
	containerStatus.load()
	pod := testStatusPod(testWaitContainerState("not-present"))
	shouldReap, _ := containerStatus.ShouldReap(pod)
	if shouldReap {
		test.Fail()
	}
}
