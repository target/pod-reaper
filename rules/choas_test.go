package rules

import (
	"os"
	"testing"
	"k8s.io/client-go/pkg/api/v1"
)

func TestChaosLoad(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CHAOS_CHANCE, "0.5")
	active, err := (&chaos{}).load()
	if !active || err != nil {
		test.Fail()
	}
}

func TestChaosFailLoad(test *testing.T) {
	os.Clearenv()
	active, err := (&chaos{}).load()
	if active || err != nil {
		test.Fail()
	}
}

func TestChaosInvalidLoad(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CHAOS_CHANCE, "not-a-number")
	active, err := (&chaos{}).load()
	if active || err == nil {
		test.Fail()
	}
}

func TestChaosShouldReap(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CHAOS_CHANCE, "1.0") // always
	chaos := chaos{}
	chaos.load()
	shouldReap, message := chaos.ShouldReap(v1.Pod{})
	if !shouldReap || message != "was flagged for chaos" {
		test.Fail()
	}
}

func TestChaosShouldNotReap(test *testing.T) {
	os.Clearenv()
	os.Setenv(ENV_CHAOS_CHANCE, "0.0") // never
	chaos := chaos{}
	chaos.load()
	shouldReap, _ := chaos.ShouldReap(v1.Pod{})
	if shouldReap {
		test.Fail()
	}
}
