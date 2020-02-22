package rules

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	v1 "k8s.io/client-go/pkg/api/v1"
)

const envChaosChance = "CHAOS_CHANCE"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func chaos(pod v1.Pod) (result, string) {
	value, active := os.LookupEnv(envChaosChance)
	if !active {
		return ignore, "not configured"
	}
	chance, err := strconv.ParseFloat(value, 64)
	if err != nil {
		panic(fmt.Errorf("failed to parse %s=%s %v", envChaosChance, value, err))
	}
	random := rand.Float64()
	if random < chance {
		return reap, fmt.Sprintf("random number is less than chaos chance %f (%f)", chance, random)
	}
	return spare, fmt.Sprintf("random number is greater than or equal chaos chance %f (%f)", chance, random)
}
