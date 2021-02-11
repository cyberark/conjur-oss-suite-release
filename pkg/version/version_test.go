package version

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
)

var versionFixtureData = []string{
	"v1.3.4",
	"v1.3.3",
	"v1.20.0",
	"v0.29.20",
	"v0.2.0",
	"v0.1.3",
}

func TestVersionFromString(t *testing.T) {

	type testCase struct {
		description string
		error       bool
		expected    *semver.Version
		input       string
	}

	testCases := []testCase{
		{
			input: "v1.1.0",
			error: false,
			expected: &semver.Version{
				Major: 1,
				Minor: 1,
				Patch: 0,
			},
			description: "correct syntax",
		},
		{
			input: "v1.0",
			error: true,
			expected: &semver.Version{
				Major: 1,
				Minor: 0,
			},
			description: "non-tri-dot semver syntax",
		},
		{
			input: "v1.0.0+suite",
			error: false,
			expected: &semver.Version{
				Major:    1,
				Minor:    0,
				Metadata: "suite",
			},
			description: "no suite version given",
		},
		{
			input: "v1.0.0+suite.1",
			error: false,
			expected: &semver.Version{
				Major:    1,
				Minor:    0,
				Metadata: "suite.1",
			},
			description: "correct syntax suite version given",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			actual, err := versionFromString(tc.input)

			if tc.error {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestSuiteIterationFromSemver(t *testing.T) {

	type testCase struct {
		description string
		input       *semver.Version
		expected    int
	}

	testCases := []testCase{
		{
			expected: 1,
			input: &semver.Version{
				Major:    1,
				Minor:    1,
				Patch:    0,
				Metadata: "suite.1",
			},
			description: "build metadata has iteration",
		},
		{
			expected: 1,
			input: &semver.Version{
				Major:    1,
				Minor:    1,
				Patch:    0,
				Metadata: "suite",
			},
			description: "build metadata has no iteration",
		},
		{
			expected: 1,
			input: &semver.Version{
				Major:    1,
				Minor:    1,
				Patch:    0,
				Metadata: "",
			},
			description: "no build metadata",
		},
		{
			expected: 7,
			input: &semver.Version{
				Major:    1,
				Minor:    1,
				Patch:    0,
				Metadata: "suite.7",
			},
			description: "build metadata has non-default iteration",
		},
		{
			expected: 72,
			input: &semver.Version{
				Major:    1,
				Minor:    1,
				Patch:    0,
				Metadata: "suite.72",
			},
			description: "build metadata has non-default multi-digit iteration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			actual := suiteIteration(tc.input)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestLatestReleaseInDir(t *testing.T) {
	latest, err := LatestReleaseInDir("testdata/latest_releases")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "testdata/latest_releases/suite_11.5.12+suite.2.yml", latest)
}

func TestLatestReleaseInDirBadDirectoryError(t *testing.T) {
	_, err := LatestReleaseInDir("doesnotexist")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"could not read releases directory doesnotexist: "+
			"open doesnotexist: no such file or directory",
	)
}

func TestLatestReleaseInDirEmptydirectoryError(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "version_test")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(tempDir)

	_, err = LatestReleaseInDir(tempDir)
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"could not find any release files in '"+tempDir+"'",
	)
}

func TestLatestReleaseInDirBadSemverReleaseFileError(t *testing.T) {
	_, err := LatestReleaseInDir("testdata/latest_releases_bad_semver")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"could not parse semver from 'suite_3.4.yml' in "+
			"testdata/latest_releases_bad_semver (3.4 is not in dotted-tri format)",
	)
}

func TestLatestReleaseInDirBadReleasePrefixFile(t *testing.T) {
	latest, err := LatestReleaseInDir("testdata/latest_releases_extra_files")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "testdata/latest_releases_extra_files/suite_2.3.4.yml", latest)
}

func TestLatestReleaseInDirNoValidFiles(t *testing.T) {
	_, err := LatestReleaseInDir("testdata/latest_releases_no_valid_files")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"Unable to find release file starting with 'suite_' in 'testdata/latest_releases_no_valid_files'",
	)
}

func TestGetRelevantVersions(t *testing.T) {
	relevantVersions, err := GetRelevantVersions(versionFixtureData, "v0.1.3", "v1.3.4")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(
		t,
		[]string{
			"v0.2.0",
			"v0.29.20",
			"v1.3.3",
			"v1.3.4",
		},
		relevantVersions)
}

func TestGetRelevantVersionsWithExistingSameVersion(t *testing.T) {
	versions := []string{
		"v1.3.4",
		"v1.3.3",
		"v1.20.0",
	}
	relevantVersions, err := GetRelevantVersions(versions, "v1.3.3", "v1.3.3")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, []string{"v1.3.3"}, relevantVersions)
}

func TestGetRelevantVersionsWithNonExistingSameVersion(t *testing.T) {
	versions := []string{
		"v1.3.4",
		"v1.3.3",
		"v1.20.0",
	}
	_, err := GetRelevantVersions(versions, "v1.3.10", "v1.3.10")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"v1.3.10 is not in available versions ([v1.3.4 v1.3.3 v1.20.0])",
	)
}

func TestGetRelevantVersionsWithOnlyLowVersionIncludesOnlyHigherVersions(t *testing.T) {
	versions := []string{
		"v1.3.4",
		"v1.3.3",
		"v1.20.0",
	}
	relevantVersions, err := GetRelevantVersions(versions, "v1.3.3", "")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, []string{"v1.3.4", "v1.20.0"}, relevantVersions)
}

func TestGetRelevantVersionsWithOnlyMaxLowVersionIncludesNoVersions(t *testing.T) {
	versions := []string{
		"v1.3.4",
		"v1.3.3",
		"v1.20.0",
	}
	relevantVersions, err := GetRelevantVersions(versions, "v1.20.0", "")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, []string{}, relevantVersions)
}

func TestGetRelevantVersionsWithLowVersionThatIsMissingIncludesHigherVersions(t *testing.T) {
	relevantVersions, err := GetRelevantVersions(versionFixtureData, "v1.2.2", "")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(
		t,
		[]string{
			"v1.3.3",
			"v1.3.4",
			"v1.20.0",
		},
		relevantVersions,
	)
}

func TestGetRelevantVersionsWithOnlyHighVersion(t *testing.T) {
	versions := []string{
		"v1.3.4",
		"v1.3.3",
		"v1.20.0",
	}
	relevantVersions, err := GetRelevantVersions(versions, "", "v1.3.3")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, []string{"v1.3.3"}, relevantVersions)
}

func TestGetRelevantVersionsWithOnlyHighVersionThatIsMissing(t *testing.T) {
	_, err := GetRelevantVersions(versionFixtureData, "", "v9.9.9")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"v9.9.9 is not in available versions ([v1.3.4 v1.3.3 v1.20.0 v0.29.20 v0.2.0 v0.1.3])",
	)
}

func TestGetRelevantVersionsWithBadVersionInList(t *testing.T) {
	versions := []string{
		"v1.3.4",
		"v1.3.3",
		"1.20",
	}

	_, err := GetRelevantVersions(versions, "v1.20.0", "v1.3.3")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err, "1.20 is not in dotted-tri format")
}

func TestGetRelevantVersionsBadHighVersion(t *testing.T) {
	_, err := GetRelevantVersions([]string{}, "123", "v1.3.3")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err, "123 is not in dotted-tri format")
}

func TestGetRelevantVersionsBadLowVersion(t *testing.T) {
	_, err := GetRelevantVersions([]string{}, "v1.2.3", "asddsf")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err, "asddsf is not in dotted-tri format")
}

func TestHighestVersion(t *testing.T) {
	highestVersion, err := HighestVersion(versionFixtureData)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, highestVersion, "v1.20.0")
}

func TestHighestVersionSingleVersionArg(t *testing.T) {
	highestVersion, err := HighestVersion([]string{"v2.3.4"})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "v2.3.4", highestVersion)
}

func TestHighestVersionNoVersionsPassedIn(t *testing.T) {
	_, err := HighestVersion([]string{})
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err, "cannot ascertain highest version - no versions provided")
}

func TestHighestVersionFirstVersionBadInArg(t *testing.T) {
	_, err := HighestVersion([]string{"abcd"})
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err, "abcd is not in dotted-tri format")
}

func TestHighestVersionBadVersionsInArg(t *testing.T) {
	_, err := HighestVersion(
		[]string{
			"1.2.3",
			"2.3.4",
			"9",
			"3.4.5",
		},
	)
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err, "9 is not in dotted-tri format")
}
