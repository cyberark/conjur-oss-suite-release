package github

import (
	"encoding/json"
	"fmt"
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
		if strings.Contains(url, "real_release_branch") {
			// We are running a test where the repo has a release branch
			return httpClient.Get("file://./testdata/release_branch_changelog.md")
		} else if strings.Contains(url, "repo_with_main_branch") {
			// We are running a test where the repo has a main branch
			return httpClient.Get("file://./testdata/main_branch_changelog.md")
		}
		return httpClient.Get("file://./testdata/simple_changelog.md")
	} else if strings.Contains(url, "branches") {
		if strings.Contains(url, "real_release_branch") {
			// We are running a test where the repo has a release branch
			return httpClient.Get("file://./testdata/branch_v3.json")
		} else if strings.Contains(url, "repo_with_main_branch/branches/main") {
			// We are running a test where the repo has a release branch
			return httpClient.Get("file://./testdata/main_v3.json")
		}
		return nil, fmt.Errorf("Branch not found")
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

	assert.Equal(t, len(releases), 4)

	expectedReleases := []ReleaseInfo{
		ReleaseInfo{
			TagName: "v0.1.1",
			Name:    "v0.1.1",
			Draft:   false,
		},
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
		"v0.1.1",
		"v0.0.5",
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
		"Get \"doesnotexist\": unsupported protocol scheme \"\"",
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

func TestGetAvailableReleasesBadSemver(t *testing.T) {

	expectedReleases := []string{
		"v1.0.6",
		"v1.0.6",
		"v1.0.5",
		"v1.0.4",
		"v1.0.2",
		"v1.0.0-rc4",
		"v1.0.0-rc1",
	}

	httpClient := generateHTTPClientWithFileSupportTransport()

	// bad_semver_releases_v3.json should skip versions with bad semver
	actualReleases, err := getAvailableReleases(
		httpClient,
		"file://./testdata/bad_semver_releases_v3.json",
	)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, expectedReleases, actualReleases)
}

func TestCollectSuiteCategories(t *testing.T) {
	mockClient := NewMockClient()

	testCases := []struct {
		description             string
		shouldIncludeChangelogs bool
		changelogVersions       int
		oldSuiteFileName        string
		newSuiteFileName        string
		categoryName            string
		releaseBranch           string
	}{
		{
			description:             "2 version diff between suite versions, but only 1 diff in changelog",
			shouldIncludeChangelogs: true,
			changelogVersions:       1,
			oldSuiteFileName:        "old_suite.yml",
			newSuiteFileName:        "new_suite.yml",
			categoryName:            "Conjur SDK",
			releaseBranch:           "",
		},
		{
			description:             "no changes between repo versions",
			shouldIncludeChangelogs: false,
			changelogVersions:       0,
			oldSuiteFileName:        "new_suite.yml",
			newSuiteFileName:        "new_suite.yml",
			categoryName:            "Conjur SDK",
			releaseBranch:           "",
		},
		{
			description:             "2 version diff between suite versions, using release branch with full changelog",
			shouldIncludeChangelogs: true,
			changelogVersions:       2,
			oldSuiteFileName:        "old_suite.yml",
			newSuiteFileName:        "new_suite.yml",
			categoryName:            "Conjur SDK",
			releaseBranch:           "real_release_branch",
		},
		{
			description:             "2 version diff between suite versions with changelog from main branch",
			shouldIncludeChangelogs: true,
			changelogVersions:       2,
			oldSuiteFileName:        "old_main_suite.yml",
			newSuiteFileName:        "new_main_suite.yml",
			categoryName:            "Conjur SDK",
			releaseBranch:           "fake_release_branch",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			repoConfig, err := generateRepoConfig(t, tc.oldSuiteFileName, tc.newSuiteFileName)
			if !assert.NoError(t, err) {
				return
			}

			actualSuiteCategories, err := CollectSuiteCategories(repoConfig, mockClient, tc.releaseBranch)
			if !assert.NoError(t, err) {
				return
			}

			assert.NotNil(t, actualSuiteCategories)

			assert.Equal(t, tc.categoryName, actualSuiteCategories[0].CategoryName)

			if tc.shouldIncludeChangelogs {
				assert.NotNil(t, actualSuiteCategories[0].Components[0].Changelogs)

				assert.Equal(t, tc.changelogVersions, len(actualSuiteCategories[0].Components[0].Changelogs))
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
