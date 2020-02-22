package rules

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/unversioned"
	v1 "k8s.io/client-go/pkg/api/v1"
)

func TestUnreadyIgnore(t *testing.T) {
	os.Unsetenv(envMaxUnready)
	reapResult, message := unready(v1.Pod{})
	assert.Equal(t, ignore, reapResult)
	assert.Equal(t, "not configured", message)
}

func TestUnreadyInvalid(t *testing.T) {
	os.Setenv(envMaxUnready, "invalid")
	defer func() {
		err := recover()
		assert.NotNil(t, err)
		assert.Regexp(t, "^failed to parse.*$", err)
	}()
	unready(v1.Pod{})
}

func TestUnreadyNoReadyTime(t *testing.T) {
	os.Setenv(envMaxUnready, "10m")
	reapResult, message := unready(v1.Pod{})
	assert.Equal(t, spare, reapResult)
	assert.Equal(t, "pod does not have a ready condition", message)
}

func TestUnready(t *testing.T) {
	tests := []struct {
		env                string
		lastTransitionTime time.Time
		readyStatus        v1.ConditionStatus
		reapResult         result
		messageRegex       string
	}{
		{
			env:                "1m",
			lastTransitionTime: time.Now().Add(-10 * time.Minute),
			readyStatus:        v1.ConditionTrue,
			reapResult:         spare,
			messageRegex:       "^pod is ready$",
		},
		{
			env:                "9m59s",
			lastTransitionTime: time.Now().Add(-10 * time.Minute),
			readyStatus:        v1.ConditionFalse,
			reapResult:         reap,
			messageRegex:       "^has been unready longer than 9m59s.*$",
		},
		{
			env:                "10m01s",
			lastTransitionTime: time.Now().Add(-10 * time.Minute),
			readyStatus:        v1.ConditionFalse,
			reapResult:         spare,
			messageRegex:       "^has been unready less than 10m1s.*$",
		},
	}
	for _, test := range tests {
		os.Setenv(envMaxUnready, test.env)
		pod := testUnreadyPod(test.lastTransitionTime, test.readyStatus)
		reapResult, message := unready(pod)
		assert.Equal(t, test.reapResult, reapResult)
		assert.Regexp(t, test.messageRegex, message)
	}
}

func testUnreadyPod(lastTransitionTime time.Time, readyStatus v1.ConditionStatus) v1.Pod {
	return v1.Pod{
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:               v1.PodReady,
					LastTransitionTime: unversioned.NewTime(lastTransitionTime),
					Reason:             "ContainersNotReady",
					Status:             readyStatus,
				},
			},
		},
	}
}
