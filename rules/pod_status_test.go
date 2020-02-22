package rules

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/client-go/pkg/api/v1"
)

func TestPodStatusIgnore(t *testing.T) {
	os.Unsetenv(envPodStatus)
	reapResult, message := podStatus(v1.Pod{})
	assert.Equal(t, ignore, reapResult)
	assert.Equal(t, notConfigured, message)
}

func TestPodStatus(t *testing.T) {
	tests := []struct {
		env        string
		podStatus  string
		reapResult result
		message    string
	}{
		{
			env:        "test",
			podStatus:  "test",
			reapResult: reap,
			message:    "has pod status 'test' in {test}",
		},
		{
			env:        "test",
			podStatus:  "other",
			reapResult: spare,
			message:    "has pod status 'other' not in {test}",
		},
		{
			env:        "test,other",
			podStatus:  "other",
			reapResult: reap,
			message:    "has pod status 'other' in {test,other}",
		},
		{
			env:        "test,other",
			podStatus:  "neither",
			reapResult: spare,
			message:    "has pod status 'neither' not in {test,other}",
		},
	}
	for _, test := range tests {
		os.Setenv(envPodStatus, test.env)
		pod := podStatusPod(test.podStatus)
		reapResult, message := podStatus(pod)
		assert.Equal(t, test.reapResult, reapResult)
		assert.Equal(t, test.message, message)
	}
}

func podStatusPod(reason string) v1.Pod {
	return v1.Pod{
		Status: v1.PodStatus{
			Reason: reason,
		},
	}
}
