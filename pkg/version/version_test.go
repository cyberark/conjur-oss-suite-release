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

func TestGetRelevantVersionsWithSameVersion(t *testing.T) {
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
