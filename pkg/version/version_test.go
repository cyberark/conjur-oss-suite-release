package version

import (
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
