package rules

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/unversioned"
	k8v1 "k8s.io/client-go/pkg/api/v1"
)

func TestDurationIgnore(t *testing.T) {
	os.Unsetenv(envMaxDuration)
	reapResult, message := duration(k8v1.Pod{})
	assert.Equal(t, ignore, reapResult)
	assert.Equal(t, "not configured", message)
}

func TestDurationInvalid(t *testing.T) {
	os.Setenv(envMaxDuration, "invalid")
	defer func() {
		err := recover()
		assert.NotNil(t, err)
		assert.Regexp(t, "^failed to parse.*$", err)
	}()
	duration(k8v1.Pod{})
}

func TestDurationNoStartTime(t *testing.T) {
	os.Setenv(envMaxDuration, "1m")
	reapResult, message := duration(k8v1.Pod{})
	assert.Equal(t, spare, reapResult)
	assert.Equal(t, "pod has no start time", message)
}

func TestDuration(t *testing.T) {
	tests := []struct {
		env          string
		startTime    time.Time
		reapResult   result
		messageRegex string
	}{
		{
			env:          "1m",
			startTime:    time.Now().Add(-2 * time.Minute),
			reapResult:   reap,
			messageRegex: "^pod running for longer than 1m.*$",
		},
		{
			env:          "3m",
			startTime:    time.Now().Add(-2 * time.Minute),
			reapResult:   spare,
			messageRegex: "^pod running for less than 3m.*$",
		},
	}
	for _, test := range tests {
		os.Setenv(envMaxDuration, test.env)
		pod := durationPod(test.startTime)
		reapResult, message := duration(pod)
		assert.Equal(t, test.reapResult, reapResult)
		assert.Regexp(t, test.messageRegex, message)
	}
}

func durationPod(startTime time.Time) k8v1.Pod {
	pod := k8v1.Pod{}
	setTime := unversioned.NewTime(startTime)
	pod.Status.StartTime = &setTime
	return pod
}
