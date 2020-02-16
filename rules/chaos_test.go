package rules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/client-go/pkg/api/v1"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}
func TestChaosShouldReap(t *testing.T) {
	tests := []struct {
		env          string
		reapResult   result
		messageRegex string
	}{
		{
			env:          "1.0",
			reapResult:   reap,
			messageRegex: "^random number .* < chaos chance .*$",
		},
		{
			env:          "0.0",
			reapResult:   spare,
			messageRegex: "^random number .* >= chaos chance .*$",
		},
	}
	chaos := chaos{}
	pod := v1.Pod{}
	for _, test := range tests {
		os.Setenv(envChaosChance, test.env)
		reapResult, message := chaos.shouldReap(pod)
		assert.Equal(t, test.reapResult, reapResult)
		assert.Regexp(t, test.messageRegex, message)
	}

	// test that we panic if the environment variable is invalid
	defer func() {
		assert.NotNil(t, recover())
	}()
	os.Setenv(envChaosChance, "not-a-number")
	chaos.shouldReap(pod)
}
