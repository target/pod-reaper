package rules

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/client-go/pkg/api/v1"
)

const envChaosChance = "CHAOS_CHANCE"

type chaos struct{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (rule *chaos) shouldReap(pod v1.Pod) (result, string) {
	value, active := os.LookupEnv(envChaosChance)
	if !active {
		return ignore, ""
	}
	chance, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logrus.WithError(err).Errorf("invalid chaos chance %s", value)
		panic(err) // the reaper is misconfigured
	}
	random := rand.Float64()
	if random < chance {
		return reap, fmt.Sprintf("random number %f < chaos chance %f", random, chance)
	}
	return spare, fmt.Sprintf("random number %f >= chaos chance %f", random, chance)
}
