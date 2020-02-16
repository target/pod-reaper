package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/client-go/pkg/api/v1"
)

var _ rule = (*testRule)(nil)

type testRule struct {
	fixedResult result
}

func (rule *testRule) shouldReap(pod v1.Pod) (result, string) {
	return rule.fixedResult, "fixed result"
}

var testReap = testRule{fixedResult: reap}
var testSpare = testRule{fixedResult: spare}
var testIgnore = testRule{fixedResult: ignore}

func TestShouldReap(t *testing.T) {
	tests := []struct {
		rules      []rule
		shouldReap bool
		reapCount  int
		spareCount int
	}{
		{
			rules:      []rule{&testReap},
			shouldReap: true,
			reapCount:  1,
			spareCount: 0,
		},
		{
			rules:      []rule{&testSpare},
			shouldReap: false,
			reapCount:  0,
			spareCount: 1,
		},
		{
			rules:      []rule{&testIgnore},
			shouldReap: false,
			reapCount:  0,
			spareCount: 0,
		},
		{
			rules:      []rule{&testReap, &testIgnore},
			shouldReap: true,
			reapCount:  1,
			spareCount: 0,
		},
		{
			rules:      []rule{&testSpare, &testIgnore},
			shouldReap: false,
			reapCount:  0,
			spareCount: 1,
		},
		{
			rules:      []rule{&testReap, &testSpare, &testIgnore},
			shouldReap: false,
			reapCount:  1,
			spareCount: 1,
		},
	}
	pod := v1.Pod{}
	for _, test := range tests {
		shouldReap, reapReasons, spareReasons := ShouldReap(pod, test.rules)
		assert.Equal(t, test.shouldReap, shouldReap, "unexpected ShouldReap result")
		assert.Equal(t, test.reapCount, len(reapReasons), "unexpected ShouldReap reapReasons count")
		assert.Equal(t, test.spareCount, len(spareReasons), "unexpected ShouldReap spareReasons count")
	}
}
