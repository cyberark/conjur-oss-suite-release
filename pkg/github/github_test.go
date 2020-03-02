package github

import (
	"encoding/json"
	"io/ioutil"
	stdlibHttp "net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	pkgHttp "github.com/cyberark/conjur-oss-suite-release/pkg/http"
)

func TestReleaseParsing(t *testing.T) {
	releaseJSON, err := ioutil.ReadFile("testdata/release_v3.json")
	if !assert.NoError(t, err) {
		return
	}

	var releaseInfo = ReleaseInfo{}
	err = json.Unmarshal(releaseJSON, &releaseInfo)
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
	releaseJSON, err := ioutil.ReadFile("testdata/releases_v3.json")
	if !assert.NoError(t, err) {
		return
	}

	var releases = []ReleaseInfo{}
	err = json.Unmarshal(releaseJSON, &releases)
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

func TestGetAvailableReleases(t *testing.T) {
	expectedReleases := []string{
		"v0.0.5",
		"v0.0.4",
		"v0.0.3",
	}

	transportWithFileSupport := &stdlibHttp.Transport{}
	transportWithFileSupport.RegisterProtocol(
		"file",
		stdlibHttp.NewFileTransport(stdlibHttp.Dir(".")),
	)

	httpClient := &pkgHttp.Client{
		&stdlibHttp.Client{
			Transport: transportWithFileSupport,
		},
		"",
	}

	actualReleases, err := getAvailableReleases(
		httpClient,
		"file://./testdata/releases_v3.json",
	)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, expectedReleases, actualReleases)
}

func TestGetAvailableReleasesFetchingProblem(t *testing.T) {
	httpClient := &pkgHttp.Client{
		&stdlibHttp.Client{},
		"",
	}

	_, err := getAvailableReleases(
		httpClient,
		"doesnotexist",
	)
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"Get doesnotexist: unsupported protocol scheme \"\"",
	)
}

func TestGetAvailableReleasesUnmarshalingProblem(t *testing.T) {
	transportWithFileSupport := &stdlibHttp.Transport{}
	transportWithFileSupport.RegisterProtocol(
		"file",
		stdlibHttp.NewFileTransport(stdlibHttp.Dir(".")),
	)

	httpClient := &pkgHttp.Client{
		&stdlibHttp.Client{
			Transport: transportWithFileSupport,
		},
		"",
	}

	// release_v3.json (vs releases_v3.json) should fail unmarshaling since it's
	// not an array
	_, err := getAvailableReleases(
		httpClient,
		"file://./testdata/release_v3.json",
	)
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"json: cannot unmarshal object into Go value of type []github.ReleaseInfo",
	)
}
