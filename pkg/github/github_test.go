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
	assert.Regexp(t, regexp.MustCompile("^# Change log"), description)
	assert.Regexp(t, regexp.MustCompile("\\(#1062\\)$"), description)

	assert.Equal(t, releaseInfo.TagName, "v1.4.2")
	assert.Equal(t, releaseInfo.Draft, false)
	assert.Equal(t, releaseInfo.Name, "v1.4.2")
}

func TestReleasesParsing(t *testing.T) {
	releaseJson, err := ioutil.ReadFile("testdata/releases_v3.json")
	if !assert.NoError(t, err) {
		return
	}

	var releases = []ReleaseInfo{}
	err = json.Unmarshal(releaseJson, &releases)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, len(releases), 3)

	expectedReleases := []ReleaseInfo{
		ReleaseInfo{
			TagName: "v0.0.5",
			Name:    "v0.0.5",
			Draft:   false,
		},
		ReleaseInfo{
			TagName: "v0.0.4",
			Name:    "v0.0.4",
			Draft:   false,
		},
		ReleaseInfo{
			TagName: "v0.0.3",
			Name:    "v0.0.3",
			Draft:   false,
		},
	}

	for index, actualRelase := range releases {
		assert.Equal(t, actualRelase.TagName, expectedReleases[index].TagName)
		assert.Equal(t, actualRelase.Name, expectedReleases[index].Name)
		assert.Equal(t, actualRelase.Draft, expectedReleases[index].Draft)
	}
}
