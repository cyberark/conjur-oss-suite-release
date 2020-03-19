// +build release_test

package main

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func callBashScript(script string) (string, error) {
	out, err := exec.Command(
		"bash",
		"-c",
		`
set -euo pipefail;
. ./executors.sh;
` + script,
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

func TestSecretless(t *testing.T) {
	t.Run("Consume conjur variable", func(t *testing.T) {
		expectedValue := "abc123"

		_, err := callBashScript(
			// language=bash
			`
exec_conjur_client conjur variable values add test-app-secrets/username ` + expectedValue,
		)
		if !assert.NoError(t, err) {
			return
		}

		out, err := callBashScript(
			// language=bash
			`
exec_app bash -exc '
export http_proxy=http://localhost:8080

curl -s "http://httpbin.org/anything" | jq -j ".headers.Authorization"
'
`)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, expectedValue, out)
	})
}
