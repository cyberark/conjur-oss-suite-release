package github

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
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

	description := releaseInfo.Description
	tagName := releaseInfo.TagName

	assert.Regexp(t, regexp.MustCompile("^# Change log"), description)
	assert.Regexp(t, regexp.MustCompile("\\(#1062\\)$"), description)

	assert.Equal(t, tagName, "v1.4.2")
}
