package rules

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

const (
	ruleChaos             = "chaos"
	envChaosChance        = "CHAOS_CHANCE"
	annotationChaosChance = annotationPrefix + "/chaos-chance"
)

var _ Rule = (*chaos)(nil)

type chaos struct {
	chance float64
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (rule *chaos) load() (bool, string, error) {
	explicit := explicitLoad(ruleChaos)
	value, hasDefault := os.LookupEnv(envChaosChance)
	if !explicit && !hasDefault {
		return false, "", nil
	}
	chance, err := strconv.ParseFloat(value, 64)
	if !explicit && err != nil {
		return false, "", fmt.Errorf("invalid chaos chance %s", err)
	}
	rule.chance = chance

	if rule.chance != 0 {
		return true, fmt.Sprintf("chaos chance %s", value), nil
	}
	return true, fmt.Sprint("chaos (no default)"), nil
}

func (rule *chaos) ShouldReap(pod v1.Pod) (bool, string) {
	chance := rule.chance
	annotationValue := pod.Annotations[annotationChaosChance]
	if annotationValue != "" {
		annotationChance, err := strconv.ParseFloat(annotationValue, 64)
		if err == nil {
			chance = annotationChance
		} else {
			logrus.Warnf("pod %s has invalid chaos chance: %s", pod.Name, err)
		}
	}

	return rand.Float64() < chance, "was flagged for chaos"
}
