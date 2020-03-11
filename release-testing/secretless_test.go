// +build release_test

package main

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"test/kv"
)

var storeClient *kv.StoreClient
func init() {
	var err error
	storeClient, err = kv.DefaultStoreClient()
	if err != nil {
		panic(err)
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

func TestSecretless(t *testing.T) {
	t.Run("Consume conjur variable", func(t *testing.T) {
		expectedValue := "abc123"

		_, err := callCommand(
			"bash",
			"-ec",
			// language=bash
			`
. ./executors.sh

exec_conjur_client conjur variable values add test-app-secrets/username ` + expectedValue,
		)
		if !assert.NoError(t, err) {
			return
		}

		out, err := callCommand(
			"bash",
			"-ec",
			// language=bash
			`
. ./executors.sh

exec_app bash -c '
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
