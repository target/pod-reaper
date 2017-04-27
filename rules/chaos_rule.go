package rules

import (
	"k8s.io/client-go/pkg/api/v1"
	"math/rand"
	"os"
	"strconv"
	"fmt"
)

type chaosRule struct {
	chance float64
}

func (rule *chaosRule) load() bool {
	value, active := os.LookupEnv("CHAOS_CHANCE")
	if !active {
		return false
	}
	chance, err := strconv.ParseFloat(value, 64)
	if err != nil {
		panic(fmt.Errorf("invalid chaos chance %s", err))
	}
	fmt.Printf("loading rule: chaos chance %f\n", chance)
	rule.chance = chance
	return true
}

func (rule *chaosRule) ShouldReap(pod v1.Pod) (bool, string) {
	return rand.Float64() < rule.chance, "was flagged for chaos"
}