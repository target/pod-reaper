package main

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/labels"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
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
}

func TestOptionsLoad(t *testing.T) {
	t.Run("invalid options", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envRunDuration, "invalid")
		_, err := loadOptions()
		assert.Error(t, err)
	})
}
