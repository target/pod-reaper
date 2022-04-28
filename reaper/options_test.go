package main

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"os"
	"testing"
	"time"

	"io/ioutil"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

func epocPlus(duration time.Duration) *metav1.Time {
	t := metav1.NewTime(time.Unix(0, 0).Add(duration))
	return &t
}
func testPodList() []v1.Pod {
	return []v1.Pod{
		{
			Status: v1.PodStatus{
				StartTime: epocPlus(2 * time.Minute),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "bearded-dragon",
				Annotations: map[string]string{"example/key": "lizard", "controller.kubernetes.io/pod-deletion-cost": "invalid"},
			},
		},
		{
			Status: v1.PodStatus{},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "nil-start-time",
				Annotations: map[string]string{"example/key": "lizard", "controller.kubernetes.io/pod-deletion-cost": "-100"},
			},
		},
		{
			Status: v1.PodStatus{
				StartTime: epocPlus(5 * time.Minute),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "expensive",
				Annotations: map[string]string{"example/key": "not-lizard", "controller.kubernetes.io/pod-deletion-cost": "500"},
			},
		},
		{
			Status: v1.PodStatus{
				StartTime: epocPlus(1 * time.Minute),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "corgi",
				Annotations: map[string]string{"example/key": "not-lizard"},
			},
		},
	}
}

func TestOptions(t *testing.T) {
	t.Run("namespace", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			os.Clearenv()
			namespace := namespace()
			assert.Equal(t, "", namespace)
		})
		t.Run("valid", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envNamespace, "test-namespace")
			namespace := namespace()
			assert.Equal(t, "test-namespace", namespace)
		})
	})
	t.Run("grace period", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			os.Clearenv()
			gracePeriod, err := gracePeriod()
			assert.NoError(t, err)
			assert.Nil(t, gracePeriod)
		})
		t.Run("valid", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envGracePeriod, "1m53s999ms")
			gracePeriod, err := gracePeriod()
			assert.NoError(t, err)
			assert.Equal(t, int64(113), *gracePeriod)
		})
		t.Run("invalid", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envGracePeriod, "invalid")
			_, err := gracePeriod()
			assert.Error(t, err)
		})
	})
	t.Run("schedule", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			os.Clearenv()
			schedule := schedule()
			assert.Equal(t, "@every 1m", schedule)
		})
	})
	t.Run("run duration", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			os.Clearenv()
			duration, err := runDuration()
			assert.NoError(t, err)
			assert.Equal(t, 0*time.Second, duration)
		})
		t.Run("invalid", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envRunDuration, "not-a-duration")
			_, err := runDuration()
			assert.Error(t, err)
		})
		t.Run("valid", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envRunDuration, "1m58s")
			duration, err := runDuration()
			assert.NoError(t, err)
			assert.Equal(t, 2*time.Minute-2*time.Second, duration)
		})
	})
	t.Run("label exclusion", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			os.Clearenv()
			exclusion, err := labelExclusion()
			assert.NoError(t, err)
			assert.Nil(t, exclusion)
		})
		t.Run("only key", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envExcludeLabelKey, "test-key")
			_, err := labelExclusion()
			assert.Error(t, err)
		})
		t.Run("only values", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envExcludeLabelValues, "test-value1,test-value2")
			_, err := labelExclusion()
			assert.Error(t, err)
		})
		t.Run("invalid key", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envExcludeLabelKey, "keys cannot have spaces")
			os.Setenv(envExcludeLabelValues, "test-value1,test-value2")
			_, err := labelExclusion()
			assert.Error(t, err)
		})
		t.Run("valid", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envExcludeLabelKey, "test-key")
			os.Setenv(envExcludeLabelValues, "test-value1,test-value2")
			exclusion, err := labelExclusion()
			assert.NoError(t, err)
			assert.NotNil(t, exclusion)
			assert.Equal(t, "test-key notin (test-value1,test-value2)", labels.NewSelector().Add(*exclusion).String())
		})
	})
	t.Run("label requirement", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			os.Clearenv()
			requirement, err := labelRequirement()
			assert.NoError(t, err)
			assert.Nil(t, requirement)
		})
		t.Run("only key", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envRequireLabelKey, "test-key")
			_, err := labelRequirement()
			assert.Error(t, err)
		})
		t.Run("only values", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envRequireLabelValues, "test-value1,test-value2")
			_, err := labelRequirement()
			assert.Error(t, err)
		})
		t.Run("invalid key", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envRequireLabelKey, "keys cannot have spaces")
			os.Setenv(envRequireLabelValues, "test-value1,test-value2")
			_, err := labelRequirement()
			assert.Error(t, err)
		})
		t.Run("valid", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envRequireLabelKey, "test-key")
			os.Setenv(envRequireLabelValues, "test-value1,test-value2")
			requirement, err := labelRequirement()
			assert.NoError(t, err)
			assert.NotNil(t, requirement)
			assert.Equal(t, "test-key in (test-value1,test-value2)", labels.NewSelector().Add(*requirement).String())
		})
	})
	t.Run("annotation requirement", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			os.Clearenv()
			requirement, err := annotationRequirement()
			assert.NoError(t, err)
			assert.Nil(t, requirement)
		})
		t.Run("only key", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envRequireAnnotationKey, "test-key")
			_, err := annotationRequirement()
			assert.Error(t, err)
		})
		t.Run("only values", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envRequireAnnotationValues, "test-value1,test-value2")
			_, err := annotationRequirement()
			assert.Error(t, err)
		})
		t.Run("invalid key", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envRequireAnnotationKey, "keys cannot have spaces")
			os.Setenv(envRequireAnnotationValues, "test-value1,test-value2")
			_, err := annotationRequirement()
			assert.Error(t, err)
		})
		t.Run("valid", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envRequireAnnotationKey, "test-key")
			os.Setenv(envRequireAnnotationValues, "test-value1,test-value2")
			requirement, err := annotationRequirement()
			assert.NoError(t, err)
			assert.NotNil(t, requirement)
			assert.Equal(t, "test-key in (test-value1,test-value2)", labels.NewSelector().Add(*requirement).String())
		})
	})
	t.Run("dry-run", func(t *testing.T) {
		t.Run("false", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envDryRun, "false")
			dryRun, err := dryRun()
			assert.NoError(t, err)
			assert.False(t, dryRun)
		})
		t.Run("true", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envDryRun, "true")
			dryRun, err := dryRun()
			assert.NoError(t, err)
			assert.True(t, dryRun)
		})
		t.Run("true by number", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envDryRun, "1")
			dryRun, err := dryRun()
			assert.NoError(t, err)
			assert.True(t, dryRun)
		})
		t.Run("invalid", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envDryRun, "outside expected values")
			_, err := dryRun()
			assert.Error(t, err)
		})
		t.Run("not set", func(t *testing.T) {
			os.Clearenv()
			dryRun, err := dryRun()
			assert.NoError(t, err)
			assert.False(t, dryRun)
		})
	})
	t.Run("max-pods", func(t *testing.T) {
		t.Run("invalid", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envMaxPods, "not a number")
			_, err := maxPods()
			assert.Error(t, err)
		})
		t.Run("negative", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envMaxPods, "-123")
			maxPods, err := maxPods()
			assert.NoError(t, err)
			assert.Equal(t, 0, maxPods)
		})
		t.Run("positive", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envMaxPods, "123")
			maxPods, err := maxPods()
			assert.NoError(t, err)
			assert.Equal(t, 123, maxPods)
		})
		t.Run("not set", func(t *testing.T) {
			os.Clearenv()
			maxPods, err := maxPods()
			assert.NoError(t, err)
			assert.Equal(t, 0, maxPods)
		})
	})
	t.Run("pod-sorting", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			os.Clearenv()
			sorter, err := podSortingStrategy()
			assert.NotNil(t, sorter)
			assert.NoError(t, err)
			subject := testPodList()
			sorter(subject)
			assert.Equal(t, testPodList(), subject)
		})
		t.Run("invalid", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envPodSortingStrategy, "not a valid sorting strategy")
			_, err := podSortingStrategy()
			assert.Error(t, err)
		})
		t.Run("random", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envPodSortingStrategy, "random")
			sorter, err := podSortingStrategy()
			assert.NotNil(t, sorter)
			assert.NoError(t, err)
			subject := testPodList()
			rand.Seed(2) // magic seed to force switch
			sorter(subject)
			assert.NotEqual(t, testPodList(), subject)
			assert.ElementsMatch(t, testPodList(), subject)
		})
		t.Run("oldest-first", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envPodSortingStrategy, "oldest-first")
			sorter, err := podSortingStrategy()
			assert.NotNil(t, sorter)
			assert.NoError(t, err)
			subject := testPodList()
			sorter(subject)
			assert.Equal(t, "corgi", subject[0].ObjectMeta.Name)
			assert.Equal(t, "bearded-dragon", subject[1].ObjectMeta.Name)
			assert.Equal(t, "expensive", subject[2].ObjectMeta.Name)
			assert.Equal(t, "nil-start-time", subject[3].ObjectMeta.Name)
			assert.ElementsMatch(t, testPodList(), subject)
		})
		t.Run("youngest-first", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envPodSortingStrategy, "youngest-first")
			sorter, err := podSortingStrategy()
			assert.NotNil(t, sorter)
			assert.NoError(t, err)
			subject := testPodList()
			sorter(subject)
			assert.Equal(t, "expensive", subject[0].ObjectMeta.Name)
			assert.Equal(t, "bearded-dragon", subject[1].ObjectMeta.Name)
			assert.Equal(t, "corgi", subject[2].ObjectMeta.Name)
			assert.Equal(t, "nil-start-time", subject[3].ObjectMeta.Name)
			assert.ElementsMatch(t, testPodList(), subject)
		})
		t.Run("pod-deletion-cost", func(t *testing.T) {
			os.Clearenv()
			os.Setenv(envPodSortingStrategy, "pod-deletion-cost")
			sorter, err := podSortingStrategy()
			assert.NotNil(t, sorter)
			assert.NoError(t, err)
			subject := testPodList()
			sorter(subject)
			assert.Equal(t, "nil-start-time", subject[0].ObjectMeta.Name)
			assert.Equal(t, "expensive", subject[3].ObjectMeta.Name)
			assert.ElementsMatch(t, testPodList(), subject)
		})
	})
}

func TestOptionsLoad(t *testing.T) {
	t.Run("invalid options", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envRunDuration, "invalid")
		_, err := loadOptions()
		assert.Error(t, err)
	})
	t.Run("no rules", func(t *testing.T) {
		os.Clearenv()
		_, err := loadOptions()
		assert.Error(t, err)
	})
	t.Run("valid", func(t *testing.T) {
		os.Clearenv()
		// ensure at least one rule loads
		os.Setenv("CHAOS_CHANCE", "1.0")
		options, err := loadOptions()
		assert.NoError(t, err)
		assert.Equal(t, "@every 1m", options.schedule)
		assert.Equal(t, 0*time.Second, options.runDuration)
		assert.Nil(t, options.labelExclusion)
		assert.Nil(t, options.labelRequirement)
	})
}
