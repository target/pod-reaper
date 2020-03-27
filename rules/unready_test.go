package rules

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testUnreadyPod(lastTransitionTime *time.Time) v1.Pod {
	pod := v1.Pod{}
	if lastTransitionTime != nil {
		setTime := metav1.NewTime(*lastTransitionTime)
		pod.Status.Conditions = []v1.PodCondition{
			v1.PodCondition{
				Type:               v1.PodReady,
				LastTransitionTime: setTime,
				Reason:             "ContainersNotReady",
				Status:             "False",
			},
		}
	}
	return pod
}

func TestUnreadyLoad(t *testing.T) {
	t.Run("load", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envMaxUnready, "30m")
		loaded, message, err := (&unready{}).load()
		assert.NoError(t, err)
		assert.Equal(t, "maximum unready 30m", message)
		assert.True(t, loaded)
	})
	t.Run("invalid time", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envMaxUnready, "not-a-time")
		loaded, message, err := (&unready{}).load()
		assert.Error(t, err)
		assert.Equal(t, "", message)
		assert.False(t, loaded)
	})
	t.Run("no load", func(t *testing.T) {
		os.Clearenv()
		loaded, message, err := (&unready{}).load()
		assert.NoError(t, err)
		assert.Equal(t, "", message)
		assert.False(t, loaded)
	})
}

func TestUnreadyShouldReap(t *testing.T) {
	t.Run("no ready time", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envMaxUnready, "10m")
		unready := unready{}
		unready.load()
		pod := testUnreadyPod(nil)
		shouldReap, _ := unready.ShouldReap(pod)
		assert.False(t, shouldReap)
	})
	t.Run("reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envMaxUnready, "9m59s")
		unready := unready{}
		unready.load()
		lastTransitionTime := time.Now().Add(-10 * time.Minute)
		pod := testUnreadyPod(&lastTransitionTime)
		shouldReap, reason := unready.ShouldReap(pod)
		assert.True(t, shouldReap)
		assert.Regexp(t, ".*has been unready.*", reason)
	})
	t.Run("no reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envMaxUnready, "10m1s")
		unready := unready{}
		unready.load()
		lastTransitionTime := time.Now().Add(-10 * time.Minute)
		pod := testUnreadyPod(&lastTransitionTime)
		shouldReap, _ := unready.ShouldReap(pod)
		assert.False(t, shouldReap)
	})
}
