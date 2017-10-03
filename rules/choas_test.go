package rules

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/v1"
)

func TestChaosLoad(t *testing.T) {
	t.Run("load", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(EnvChaosChance, "0.5")
		loaded, err := (&chaos{}).load()
		assert.NoError(t, err)
		assert.True(t, loaded)
	})
	t.Run("no load", func(t *testing.T) {
		os.Clearenv()
		loaded, err := (&chaos{}).load()
		assert.NoError(t, err)
		assert.False(t, loaded)
	})
	t.Run("invalid chance", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(EnvChaosChance, "not-a-number")
		loaded, err := (&chaos{}).load()
		assert.Error(t, err)
		assert.False(t, loaded)
	})
}

func TestChaosShouldReap(t *testing.T) {
	t.Run("reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(EnvChaosChance, "1.0") // always
		chaos := chaos{}
		chaos.load()
		shouldReap, message := chaos.ShouldReap(v1.Pod{})
		assert.True(t, shouldReap)
		assert.Equal(t, "was falgged for chaos", message)
	})
	t.Run("no reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(EnvChaosChance, "0.0") // never
		chaos := chaos{}
		chaos.load()
		shouldReap, _ := chaos.ShouldReap(v1.Pod{})
		assert.False(t, shouldReap)
	})
}
