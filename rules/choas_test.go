package rules

import (
	"os"
	"testing"
	"k8s.io/client-go/pkg/api/v1"
)

func TestChaosLoad(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CHAOS_CHANCE, "0.5")
	loaded, err := (&chaos{}).load()
	if !loaded {
		test.Error("not loaded")
	}
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
}

func TestChaosFailLoad(test *testing.T) {
	os.Clearenv()
	loaded, err := (&chaos{}).load()
	if loaded {
		test.Error("loaded")
	}
	if err != nil {
		test.Errorf("ERROR: %s", err)
	}
}

func TestChaosInvalidLoad(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CHAOS_CHANCE, "not-a-number")
	loaded, err := (&chaos{}).load()
	if loaded {
		test.Error("loaded")
	}
	if err == nil {
		test.Error("expected error")
	}
}

func TestChaosShouldReap(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CHAOS_CHANCE, "1.0") // always
	chaos := chaos{}
	chaos.load()
	shouldReap, message := chaos.ShouldReap(v1.Pod{})
	if !shouldReap {
		test.Error("should not reap")
	}
	if message != "was flagged for chaos" {
		test.Errorf("EXPECTED: \"was flagged for chaos\" ACTUAL: %s", message)
	}
}

func TestChaosShouldNotReap(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CHAOS_CHANCE, "0.0") // never
	chaos := chaos{}
	chaos.load()
	shouldReap, _ := chaos.ShouldReap(v1.Pod{})
	if shouldReap {
		test.Error("should reap")
	}
}
