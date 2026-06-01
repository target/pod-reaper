package main

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/target/pod-reaper/rules"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func init() {
	logrus.SetOutput(ioutil.Discard) // disable logging during tests
}

// === Test Helpers ===

// loadRulesForTest loads rules using chaos chance to control reaping behavior
// chaosChance "1.0" = always reap, "0.0" = never reap
func loadRulesForTest(chaosChance string) rules.Rules {
	os.Clearenv()
	os.Setenv("CHAOS_CHANCE", chaosChance)
	r, _ := rules.LoadRules()
	return r
}

// createTestPod creates a pod with the given name and namespace
func createTestPod(name, namespace string, startTime *time.Time) v1.Pod {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if startTime != nil {
		setTime := metav1.NewTime(*startTime)
		pod.Status.StartTime = &setTime
	}
	return pod
}

// createTestReaper creates a reaper with a fake clientset containing the provided pods
func createTestReaper(opts options, pods ...v1.Pod) reaper {
	objects := make([]runtime.Object, len(pods))
	for i := range pods {
		objects[i] = &pods[i]
	}
	return reaper{
		clientSet: fake.NewSimpleClientset(objects...),
		options:   opts,
	}
}

// minimalOptions creates minimal valid options for testing
// Pass chaosChance "0.0" for no reaping, "1.0" for always reap
func minimalOptions(chaosChance string) options {
	return options{
		namespace:          "default",
		schedule:           "@every 1m",
		podSortingStrategy: defaultSort,
		rules:              loadRulesForTest(chaosChance),
	}
}

func TestReaperFilter(t *testing.T) {
	pods := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "bearded-dragon",
				Annotations: map[string]string{"example/key": "lizard"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "corgi",
				Annotations: map[string]string{"example/key": "not-lizard"},
			},
		},
	}
	annotationRequirement, _ := labels.NewRequirement("example/key", selection.In, []string{"lizard"})
	reaper := reaper{
		options: options{
			annotationRequirement: annotationRequirement,
		},
	}
	filteredPods := filter(reaper, pods...)
	assert.Equal(t, 1, len(filteredPods))
	assert.Equal(t, "bearded-dragon", filteredPods[0].ObjectMeta.Name)
}

// === getPods Tests ===

func TestGetPods(t *testing.T) {
	t.Run("basic list", func(t *testing.T) {
		startTime := time.Now()
		pods := []v1.Pod{
			createTestPod("pod-1", "default", &startTime),
			createTestPod("pod-2", "default", &startTime),
			createTestPod("pod-3", "default", &startTime),
		}
		opts := minimalOptions("0.0")
		r := createTestReaper(opts, pods...)

		podList := r.getPods()
		assert.Equal(t, 3, len(podList.Items))
	})

	t.Run("namespace filtering", func(t *testing.T) {
		startTime := time.Now()
		pods := []v1.Pod{
			createTestPod("pod-1", "default", &startTime),
			createTestPod("pod-2", "kube-system", &startTime),
			createTestPod("pod-3", "default", &startTime),
		}
		opts := minimalOptions("0.0")
		opts.namespace = "default"
		r := createTestReaper(opts, pods...)

		podList := r.getPods()
		assert.Equal(t, 2, len(podList.Items))
		for _, pod := range podList.Items {
			assert.Equal(t, "default", pod.Namespace)
		}
	})

	t.Run("all namespaces", func(t *testing.T) {
		startTime := time.Now()
		pods := []v1.Pod{
			createTestPod("pod-1", "default", &startTime),
			createTestPod("pod-2", "kube-system", &startTime),
		}
		opts := minimalOptions("0.0")
		opts.namespace = "" // empty = all namespaces
		r := createTestReaper(opts, pods...)

		podList := r.getPods()
		assert.Equal(t, 2, len(podList.Items))
	})

	t.Run("label exclusion", func(t *testing.T) {
		startTime := time.Now()
		excludedPod := createTestPod("excluded-pod", "default", &startTime)
		excludedPod.Labels = map[string]string{"exclude": "true"}
		includedPod := createTestPod("included-pod", "default", &startTime)
		includedPod.Labels = map[string]string{"exclude": "false"}

		opts := minimalOptions("0.0")
		exclusion, _ := labels.NewRequirement("exclude", selection.NotIn, []string{"true"})
		opts.labelExclusion = exclusion
		r := createTestReaper(opts, excludedPod, includedPod)

		podList := r.getPods()
		assert.Equal(t, 1, len(podList.Items))
		assert.Equal(t, "included-pod", podList.Items[0].Name)
	})

	t.Run("label requirement", func(t *testing.T) {
		startTime := time.Now()
		matchingPod := createTestPod("matching-pod", "default", &startTime)
		matchingPod.Labels = map[string]string{"app": "target"}
		nonMatchingPod := createTestPod("non-matching-pod", "default", &startTime)
		nonMatchingPod.Labels = map[string]string{"app": "other"}

		opts := minimalOptions("0.0")
		requirement, _ := labels.NewRequirement("app", selection.In, []string{"target"})
		opts.labelRequirement = requirement
		r := createTestReaper(opts, matchingPod, nonMatchingPod)

		podList := r.getPods()
		assert.Equal(t, 1, len(podList.Items))
		assert.Equal(t, "matching-pod", podList.Items[0].Name)
	})

	t.Run("annotation filter", func(t *testing.T) {
		startTime := time.Now()
		matchingPod := createTestPod("matching-pod", "default", &startTime)
		matchingPod.Annotations = map[string]string{"reap": "true"}
		nonMatchingPod := createTestPod("non-matching-pod", "default", &startTime)
		nonMatchingPod.Annotations = map[string]string{"reap": "false"}

		opts := minimalOptions("0.0")
		requirement, _ := labels.NewRequirement("reap", selection.In, []string{"true"})
		opts.annotationRequirement = requirement
		r := createTestReaper(opts, matchingPod, nonMatchingPod)

		podList := r.getPods()
		assert.Equal(t, 1, len(podList.Items))
		assert.Equal(t, "matching-pod", podList.Items[0].Name)
	})

	t.Run("sorting applied - oldest first", func(t *testing.T) {
		now := time.Now()
		oldTime := now.Add(-10 * time.Minute)
		newTime := now.Add(-1 * time.Minute)

		oldPod := createTestPod("old-pod", "default", &oldTime)
		newPod := createTestPod("new-pod", "default", &newTime)

		opts := minimalOptions("0.0")
		opts.podSortingStrategy = oldestFirstSort
		r := createTestReaper(opts, newPod, oldPod) // insert in wrong order

		podList := r.getPods()
		assert.Equal(t, 2, len(podList.Items))
		assert.Equal(t, "old-pod", podList.Items[0].Name) // oldest should be first
	})

	t.Run("list error panics", func(t *testing.T) {
		fakeClient := fake.NewSimpleClientset()
		fakeClient.PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("simulated API error")
		})

		r := reaper{
			clientSet: fakeClient,
			options:   minimalOptions("0.0"),
		}

		assert.Panics(t, func() {
			r.getPods()
		})
	})
}

// === reapPod Tests ===

func TestReapPod(t *testing.T) {
	t.Run("dry run skips deletion", func(t *testing.T) {
		startTime := time.Now()
		pod := createTestPod("test-pod", "default", &startTime)
		opts := minimalOptions("0.0")
		opts.dryRun = true
		r := createTestReaper(opts, pod)

		r.reapPod(pod, []string{"test reason"}, 0)

		// Verify pod still exists (not deleted)
		result, err := r.clientSet.CoreV1().Pods("default").Get(context.TODO(), "test-pod", metav1.GetOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "test-pod", result.Name)
	})

	t.Run("maxPods exceeded skips deletion", func(t *testing.T) {
		startTime := time.Now()
		pod := createTestPod("test-pod", "default", &startTime)
		opts := minimalOptions("0.0")
		opts.maxPods = 2
		r := createTestReaper(opts, pod)

		r.reapPod(pod, []string{"test reason"}, 2) // reapedPods >= maxPods

		// Verify pod still exists (not deleted)
		result, err := r.clientSet.CoreV1().Pods("default").Get(context.TODO(), "test-pod", metav1.GetOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "test-pod", result.Name)
	})

	t.Run("delete path", func(t *testing.T) {
		startTime := time.Now()
		pod := createTestPod("test-pod", "default", &startTime)
		opts := minimalOptions("0.0")
		opts.evict = false
		r := createTestReaper(opts, pod)

		r.reapPod(pod, []string{"test reason"}, 0)

		// Verify pod was deleted
		_, err := r.clientSet.CoreV1().Pods("default").Get(context.TODO(), "test-pod", metav1.GetOptions{})
		assert.Error(t, err) // should be NotFound
	})

	t.Run("evict path", func(t *testing.T) {
		startTime := time.Now()
		pod := createTestPod("test-pod", "default", &startTime)
		opts := minimalOptions("0.0")
		opts.evict = true
		r := createTestReaper(opts, pod)

		// Note: fake client may not fully support eviction, but we can verify no panic
		r.reapPod(pod, []string{"test reason"}, 0)
		// Just verifying it doesn't panic is sufficient for eviction path
	})

	t.Run("grace period used", func(t *testing.T) {
		startTime := time.Now()
		pod := createTestPod("test-pod", "default", &startTime)
		gracePeriod := int64(30)
		opts := minimalOptions("0.0")
		opts.gracePeriod = &gracePeriod

		fakeClient := fake.NewSimpleClientset(&pod)
		var capturedOptions metav1.DeleteOptions
		fakeClient.PrependReactor("delete", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
			deleteAction := action.(k8stesting.DeleteAction)
			capturedOptions = deleteAction.GetDeleteOptions()
			return false, nil, nil // let the fake client handle it
		})

		r := reaper{
			clientSet: fakeClient,
			options:   opts,
		}

		r.reapPod(pod, []string{"test reason"}, 0)

		assert.NotNil(t, capturedOptions.GracePeriodSeconds)
		assert.Equal(t, int64(30), *capturedOptions.GracePeriodSeconds)
	})

	t.Run("delete error is logged but does not panic", func(t *testing.T) {
		startTime := time.Now()
		pod := createTestPod("test-pod", "default", &startTime)

		fakeClient := fake.NewSimpleClientset(&pod)
		fakeClient.PrependReactor("delete", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("simulated delete error")
		})

		opts := minimalOptions("0.0")
		r := reaper{
			clientSet: fakeClient,
			options:   opts,
		}

		// Should not panic, just log the error
		assert.NotPanics(t, func() {
			r.reapPod(pod, []string{"test reason"}, 0)
		})
	})
}

// === scytheCycle Tests ===

func TestScytheCycle(t *testing.T) {
	t.Run("no pods", func(t *testing.T) {
		opts := minimalOptions("0.0")
		r := createTestReaper(opts)

		assert.NotPanics(t, func() {
			r.scytheCycle()
		})
	})

	t.Run("no matching rules - no deletions", func(t *testing.T) {
		startTime := time.Now()
		pod1 := createTestPod("pod-1", "default", &startTime)
		pod2 := createTestPod("pod-2", "default", &startTime)

		opts := minimalOptions("0.0") // chaos chance 0.0 = never reap
		r := createTestReaper(opts, pod1, pod2)

		r.scytheCycle()

		// Verify pods still exist
		result, _ := r.clientSet.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		assert.Equal(t, 2, len(result.Items))
	})

	t.Run("all pods match - all deleted", func(t *testing.T) {
		startTime := time.Now()
		pod1 := createTestPod("pod-1", "default", &startTime)
		pod2 := createTestPod("pod-2", "default", &startTime)

		opts := minimalOptions("1.0") // chaos chance 1.0 = always reap
		r := createTestReaper(opts, pod1, pod2)

		r.scytheCycle()

		// Verify pods were deleted
		result, _ := r.clientSet.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		assert.Equal(t, 0, len(result.Items))
	})

	t.Run("maxPods limits deletions", func(t *testing.T) {
		startTime := time.Now()
		pod1 := createTestPod("pod-1", "default", &startTime)
		pod2 := createTestPod("pod-2", "default", &startTime)
		pod3 := createTestPod("pod-3", "default", &startTime)

		opts := minimalOptions("1.0") // chaos chance 1.0 = always reap
		opts.maxPods = 2
		r := createTestReaper(opts, pod1, pod2, pod3)

		r.scytheCycle()

		// Only 2 should be deleted, 1 should remain
		result, _ := r.clientSet.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		assert.Equal(t, 1, len(result.Items))
	})
}

// === harvest Tests ===

func TestHarvest(t *testing.T) {
	t.Run("runs for duration", func(t *testing.T) {
		opts := minimalOptions("0.0")
		opts.schedule = "@every 10ms"
		opts.runDuration = 50 * time.Millisecond
		r := createTestReaper(opts)

		start := time.Now()
		r.harvest()
		elapsed := time.Since(start)

		assert.True(t, elapsed >= 50*time.Millisecond, "should run at least 50ms")
		assert.True(t, elapsed < 200*time.Millisecond, "should not run too long")
	})

	t.Run("invalid schedule panics", func(t *testing.T) {
		opts := minimalOptions("0.0")
		opts.schedule = "invalid-cron-expression"
		opts.runDuration = 50 * time.Millisecond
		r := createTestReaper(opts)

		assert.Panics(t, func() {
			r.harvest()
		})
	})
}
