package github

import (
	"encoding/json"
	"io/ioutil"
	stdlibHttp "net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	pkgHttp "github.com/cyberark/conjur-oss-suite-release/pkg/http"
	"github.com/cyberark/conjur-oss-suite-release/pkg/repositories"
)

type MockClient struct {
	*stdlibHttp.Client
	AuthToken string
}

// NewClient creates a Client with an initialized parent stdlibHttp.Client
// object
func NewMockClient() *MockClient {
	return &MockClient{
		&stdlibHttp.Client{},
		"",
	}
}

func (client *MockClient) Get(url string) ([]byte, error) {
	httpClient := generateHTTPClientWithFileSupportTransport()

	if strings.Contains(url, "compare") {
		return httpClient.Get("file://./testdata/compare_v3.json")
	} else if strings.Contains(url, "CHANGELOG") {
		return httpClient.Get("file://./testdata/simple_changelog.md")
	}
	// Return local data only
	return httpClient.Get("file://./testdata/releases_v3.json")
}

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

func Test_comparisonFromURL(t *testing.T) {
	expectedComparison := &ComparisonInfo{
		URL:     "https://github.com/octocat/Hello-World/compare/master...topic",
		AheadBy: 1,
	}

	httpClient := generateHTTPClientWithFileSupportTransport()

	actualComparison, err := comparisonFromURL(
		httpClient,
		"file://./testdata/compare_v3.json",
	)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, expectedComparison, actualComparison)
}

func TestGetAvailableReleases(t *testing.T) {
	expectedReleases := []string{
		"v0.0.5",
		"v0.0.4",
		"v0.0.3",
	}

	httpClient := generateHTTPClientWithFileSupportTransport()

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
	httpClient := generateHTTPClientWithFileSupportTransport()

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
func TestCollectSuiteCategories(t *testing.T) {
	mockClient := NewMockClient()

	testCases := []struct {
		description             string
		shouldIncludeChangelogs bool
		oldSuiteFileName        string
		newSuiteFileName        string
		categoryName            string
	}{
		{
			description:             "changes between repo versions",
			shouldIncludeChangelogs: true,
			oldSuiteFileName:        "old_suite.yml",
			newSuiteFileName:        "new_suite.yml",
			categoryName:            "Conjur SDK",
		},
		{
			description:             "no changes between repo versions",
			shouldIncludeChangelogs: false,
			oldSuiteFileName:        "new_suite.yml",
			newSuiteFileName:        "new_suite.yml",
			categoryName:            "Conjur SDK",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			repoConfig, err := generateRepoConfig(t, tc.oldSuiteFileName, tc.newSuiteFileName)
			if !assert.NoError(t, err) {
				return
			}

			actualSuiteCategories, err := CollectSuiteCategories(repoConfig, mockClient)

			assert.NotNil(t, actualSuiteCategories)

			assert.Equal(t, tc.categoryName, actualSuiteCategories[0].CategoryName)

			if tc.shouldIncludeChangelogs {
				assert.NotNil(t, actualSuiteCategories[0].Components[0].Changelogs)
			}
		})
	}
}

func generateHTTPClientWithFileSupportTransport() *pkgHttp.Client {
	transportWithFileSupport := &stdlibHttp.Transport{}
	transportWithFileSupport.RegisterProtocol(
		"file",
		stdlibHttp.NewFileTransport(stdlibHttp.Dir(".")),
	)

	return &pkgHttp.Client{
		Client: &stdlibHttp.Client{
			Transport: transportWithFileSupport,
		},
		AuthToken: "",
	}
}

func generateRepoConfig(t *testing.T,
	oldSuiteFileName string,
	newSuiteFileName string) (repositories.Config, error) {
	previousConfig := repositories.Config{}

	// Construct a path to our test repositories yaml
	thisDir, err := os.Getwd()
	if !assert.NoError(t, err) {
		return previousConfig, err
	}

	oldSuiteFilePath := filepath.Join(thisDir, "testdata", oldSuiteFileName)
	previousConfig, err = repositories.NewConfig(oldSuiteFilePath)
	if !assert.NoError(t, err) {
		return previousConfig, err
	}

	if newSuiteFileName != "" {
		testFilepath := filepath.Join(thisDir, "testdata", newSuiteFileName)
		updatedConfig, err := repositories.NewConfig(testFilepath)
		if !assert.NoError(t, err) {
			return previousConfig, err
		}
		updatedConfig.SetBaselineRepoVersions(&previousConfig)

		return updatedConfig, nil
	}

	return previousConfig, nil
}
