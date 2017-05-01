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
	loaded, err := (&containerStatus{}).load()
	if !loaded {
		test.Error("not loaded")
	}
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
}

func TestStatusLoadMultiple(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CONTAINER_STATUS, "test-status,another-status")
	containerStatus := containerStatus{}
	containerStatus.load()
	statuses := containerStatus.reapStatuses
	if len(statuses) != 2 {
		test.Errorf("EXPECTED: 2 ACTUAL: %d", len(statuses))
	}
	if statuses[0] != "test-status" {
		test.Errorf("EXPECTED: \"test-status\" ACTUAL: %s", statuses[0])
	}
	if statuses[1] != "another-status" {
		test.Errorf("EXPECTED: \"another-status\" ACTUAL: %s", statuses[1])
	}
}

func TestStatusFailLoad(test *testing.T) {
	os.Clearenv()
	loaded, err := (&containerStatus{}).load()
	if loaded {
		test.Error("loaded")
	}
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
}



func TestStatusWaiting(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CONTAINER_STATUS, "test-status,another-status")
	containerStatus := containerStatus{}
	containerStatus.load()
	pod := testStatusPod(testWaitContainerState("test-status"))
	shouldReap, reason := containerStatus.ShouldReap(pod)
	if !shouldReap {
		test.Error("should not reap")
	}
	if !strings.Contains(reason, "test-status") {
		test.Errorf("EXPECTED: \"test-status\" CONTAINED IN: %s", reason)
	}
}

func TestStatusTerminated(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CONTAINER_STATUS, "test-status,another-status")
	containerStatus := containerStatus{}
	containerStatus.load()
	pod := testStatusPod(testTerminatedContainerState("another-status"))
	shouldReap, reason := containerStatus.ShouldReap(pod)
	if !shouldReap {
		test.Error("should not reap")
	}
	if !strings.Contains(reason, "another-status") {
		test.Errorf("EXPECTED: \"another-status\" CONTAINED IN: %s", reason)
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
		test.Error("should reap")
	}
}
