package rules

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/client-go/pkg/api/v1"
)

func TestChaosIgnore(t *testing.T) {
	os.Unsetenv(envChaosChance)
	reapResult, message := chaos(v1.Pod{})
	assert.Equal(t, ignore, reapResult)
	assert.Equal(t, "not configured", message)
}

func TestChaosInvalid(t *testing.T) {
	os.Setenv(envChaosChance, "not-a-number")
	defer func(){
		err := recover()
		assert.NotNil(t, err)
		assert.Regexp(t, "^failed to parse.*$", err)
	}()
	chaos(v1.Pod{})
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
	for _, test := range tests {
		os.Setenv(envChaosChance, test.env)
		reapResult, message := chaos(v1.Pod{})
		assert.Equal(t, test.reapResult, reapResult)
		assert.Regexp(t, test.messageRegex, message)
	}
}
