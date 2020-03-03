package main

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

var shouldRun bool
func init()  {
	v, ok := os.LookupEnv("RELEASE_TESTS")
	shouldRun = ok && v == "1"
}
func checkShouldRun(t *testing.T) {
	if !shouldRun {
		t.Skip("skipping release tests")
	}
}


func callCommand(name string, args ...string) (string, error) {
	out, err := exec.Command(
		name,
		args...
	).Output()

	outString := string(out)
	if err != nil {
		if exitErrr, ok := err.(*exec.ExitError); ok {
			return outString, errors.New(string(exitErrr.Stderr))
		}

		return outString, err
	}

	return outString, nil
}

func TestKubernetesAuthenticator(t *testing.T) {
	checkShouldRun(t)

	t.Run("Retrieve conjur variable", func(t *testing.T) {
		expectedValue := "meow meow meow"
		out, err := callCommand("bash", "-c", `
. ./store.sh

kubectl exec "$(get_val APP_POD_NAME)" -c app -- bash -xce '
export CONJUR_AUTHN_TOKEN=$(cat /run/conjur/access-token | base64)
conjur variable value test-app-secrets/username
'
`)

		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, expectedValue, out)
	})
}
