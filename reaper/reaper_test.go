package main

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"testing"
)

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
