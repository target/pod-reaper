package rules

import (
	"k8s.io/client-go/pkg/api/v1"
	"math/rand"
	"os"
	"strconv"
	"fmt"
	"time"
)

const ENV_CHAOS_CHANCE = "CHAOS_CHANCE"

type chaos struct {
	chance float64
	random *rand.Rand
}

func createRand() *rand.Rand {
	seed := time.Now().UnixNano() // no need for crypto-level randomness
	return rand.New(rand.NewSource(seed))
}

func (rule *chaos) load() (bool, error) {
	value, active := os.LookupEnv(ENV_CHAOS_CHANCE)
	if !active {
		return false, nil
	}
	chance, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false, fmt.Errorf("invalid chaos chance %s", err)
	}
	rule.random = createRand()
	rule.chance = chance
	return true, nil
}

func (rule *chaos) ShouldReap(pod v1.Pod) (bool, string) {
	return rule.random.Float64() < rule.chance, "was flagged for chaos"
}