package rules

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func TestChaosLoad(t *testing.T) {
	t.Run("load", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envChaosChance, "0.5")
		loaded, message, err := (&chaos{}).load()
		assert.NoError(t, err)
		assert.Equal(t, "chaos chance 0.5", message)
		assert.True(t, loaded)
	})
	t.Run("no load", func(t *testing.T) {
		os.Clearenv()
		loaded, message, err := (&chaos{}).load()
		assert.NoError(t, err)
		assert.Equal(t, "", message)
		assert.False(t, loaded)
	})
	t.Run("invalid chance", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envChaosChance, "not-a-number")
		loaded, message, err := (&chaos{}).load()
		assert.Error(t, err)
		assert.Equal(t, "", message)
		assert.False(t, loaded)
	})
	t.Run("negative chance loads successfully", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envChaosChance, "-0.5")
		c := chaos{}
		loaded, message, err := c.load()
		assert.NoError(t, err)
		assert.True(t, loaded)
		assert.Equal(t, "chaos chance -0.5", message)
		assert.Equal(t, -0.5, c.chance)
	})
	t.Run("chance above 1 loads successfully", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envChaosChance, "2.0")
		c := chaos{}
		loaded, message, err := c.load()
		assert.NoError(t, err)
		assert.True(t, loaded)
		assert.Equal(t, "chaos chance 2.0", message)
		assert.Equal(t, 2.0, c.chance)
	})
	t.Run("whitespace causes parse error", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envChaosChance, " 0.5 ")
		loaded, message, err := (&chaos{}).load()
		assert.Error(t, err)
		assert.Equal(t, "", message)
		assert.False(t, loaded)
	})
}

func TestChaosShouldReap(t *testing.T) {
	t.Run("reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envChaosChance, "1.0") // always
		chaos := chaos{}
		chaos.load()
		shouldReap, message := chaos.ShouldReap(v1.Pod{})
		assert.True(t, shouldReap)
		assert.Equal(t, "was flagged for chaos", message)
	})
	t.Run("no reap", func(t *testing.T) {
		os.Clearenv()
		os.Setenv(envChaosChance, "0.0") // never
		chaos := chaos{}
		chaos.load()
		shouldReap, _ := chaos.ShouldReap(v1.Pod{})
		assert.False(t, shouldReap)
	})
	t.Run("negative chance never reaps", func(t *testing.T) {
		c := chaos{chance: -0.5}
		// rand.Float64() returns [0.0, 1.0), so it's always >= -0.5
		for i := 0; i < 100; i++ {
			shouldReap, _ := c.ShouldReap(v1.Pod{})
			assert.False(t, shouldReap)
		}
	})
	t.Run("chance above 1 always reaps", func(t *testing.T) {
		c := chaos{chance: 2.0}
		// rand.Float64() returns [0.0, 1.0), so it's always < 2.0
		for i := 0; i < 100; i++ {
			shouldReap, _ := c.ShouldReap(v1.Pod{})
			assert.True(t, shouldReap)
		}
	})
	t.Run("NaN chance never reaps", func(t *testing.T) {
		c := chaos{chance: math.NaN()}
		// Any comparison with NaN returns false
		for i := 0; i < 100; i++ {
			shouldReap, _ := c.ShouldReap(v1.Pod{})
			assert.False(t, shouldReap)
		}
	})
	t.Run("positive Inf chance always reaps", func(t *testing.T) {
		c := chaos{chance: math.Inf(1)}
		// rand.Float64() returns [0.0, 1.0), which is always < +Inf
		for i := 0; i < 100; i++ {
			shouldReap, _ := c.ShouldReap(v1.Pod{})
			assert.True(t, shouldReap)
		}
	})
	t.Run("negative Inf chance never reaps", func(t *testing.T) {
		c := chaos{chance: math.Inf(-1)}
		// rand.Float64() returns [0.0, 1.0), which is never < -Inf
		for i := 0; i < 100; i++ {
			shouldReap, _ := c.ShouldReap(v1.Pod{})
			assert.False(t, shouldReap)
		}
	})
}
