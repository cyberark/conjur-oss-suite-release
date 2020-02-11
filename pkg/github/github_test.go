package github

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReleaseParsing(t *testing.T) {
	releaseJson, err := ioutil.ReadFile("testdata/release_v3.json")
	if !assert.NoError(t, err) {
		return
	}

	var releaseInfo = ReleaseInfo{}
	err = json.Unmarshal(releaseJson, &releaseInfo)
	if !assert.NoError(t, err) {
		return
	}

	tagName := releaseInfo.TagName

	assert.Equal(t, tagName, "v1.4.2")
}
