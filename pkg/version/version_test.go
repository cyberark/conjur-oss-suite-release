package version

import (
	"io/ioutil"
	"os"
	"testing"

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

func TestLatestReleaseInDir(t *testing.T) {
	latest, err := LatestReleaseInDir("testdata/latest_releases")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "testdata/latest_releases/suite_11.5.12.yml", latest)
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
	_, err := LatestReleaseInDir("testdata/latest_releases_extra_files")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"found non-release prefix ('suite_') file 'notsuite_5.4.3.yml' "+
			"in 'testdata/latest_releases_extra_files'",
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

func TestGetRelevantVersionsWithSwappedHighAndLowVersions(t *testing.T) {
	relevantVersions, err := GetRelevantVersions(versionFixtureData, "v1.3.4", "v0.1.3")
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

func TestGetRelevantVersionsWithOnlyLowVersion(t *testing.T) {
	versions := []string{
		"v1.3.4",
		"v1.3.3",
		"v1.20.0",
	}
	relevantVersions, err := GetRelevantVersions(versions, "v1.3.3", "")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, []string{"v1.3.3"}, relevantVersions)
}

func TestGetRelevantVersionsWithOnlyLowVersionThatIsMissing(t *testing.T) {
	_, err := GetRelevantVersions(versionFixtureData, "v0.0.0", "")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"v0.0.0 is not in available versions ([v1.3.4 v1.3.3 v1.20.0 v0.29.20 v0.2.0 v0.1.3])",
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

func TestGetRelevantVersionsWithNoVersionsWithinRange(t *testing.T) {
	_, err := GetRelevantVersions(versionFixtureData, "v98.99.99", "v99.99.99")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"could not find relevant versions for range v98.99.99 -> v99.99.99 "+
			"in available versions ([v1.3.4 v1.3.3 v1.20.0 v0.29.20 v0.2.0 v0.1.3])",
	)
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

	assert.Equal(t, highestVersion, "v2.3.4")
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
