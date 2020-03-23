package version

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/coreos/go-semver/semver"
)

// ReleasesPrefix denotes the expected prefix for all release files
const ReleasesPrefix = "suite_"

func versionFromString(versionStr string) (*semver.Version, error) {
	// Strip the 'v' from the beginning, if present
	return semver.NewVersion(strings.TrimPrefix(versionStr, "v"))
}

// LatestReleaseInDir returns the file matching the highest semver for a
// group of files in the specified `releasesDir`.
func LatestReleaseInDir(releasesDir string) (string, error) {
	files, err := ioutil.ReadDir(releasesDir)
	if err != nil {
		return "", fmt.Errorf(
			"could not read releases directory %s: %s",
			releasesDir,
			err,
		)
	}

	if len(files) == 0 {
		return "", fmt.Errorf(
			"could not find any release files in '%s'",
			releasesDir,
		)
	}

	highestVersion, _ := versionFromString("0.0.0")
	highestReleaseFile := files[0].Name()
	for _, file := range files {
		filename := file.Name()

		if !strings.HasPrefix(filename, ReleasesPrefix) {
			return "", fmt.Errorf(
				"found non-release prefix ('%s') file '%s' in '%s'",
				ReleasesPrefix,
				filename,
				releasesDir,
			)
		}

		// Turns `suite_x.y.z.yml` into `x.y.z`
		versionText := strings.Replace(
			strings.TrimSuffix(filename, filepath.Ext(filename)),
			ReleasesPrefix,
			"",
			1,
		)

		version, err := versionFromString(versionText)
		if err != nil {
			return "", fmt.Errorf(
				"could not parse semver from '%s' in %s (%s)",
				filename,
				releasesDir,
				err,
			)
		}

		if highestVersion.LessThan(*version) {
			highestVersion = version
			highestReleaseFile = filename
		}
	}

	highestReleasePath := filepath.Join(releasesDir, highestReleaseFile)

	return highestReleasePath, nil
}

// GetRelevantVersions sorts and returns the list of versions from highest
// (included) to the lowest (excluded). The method auto-detects what's the
// lower and what's the higher range bound.
func GetRelevantVersions(availVersionsStr []string,
	startVersionStr string,
	endVersionStr string) ([]string, error) {

	// Allow specifying `""` as a low version
	if startVersionStr == "" {
		startVersionStr = endVersionStr
	}

	// Allow specifying `""` as a high version
	if endVersionStr == "" {
		endVersionStr = startVersionStr
	}

	// Parse the higher limit version from the provided string
	highVersion, err := versionFromString(endVersionStr)
	if err != nil {
		return nil, err
	}

	// Parse the lower limit version from the provided string
	lowVersion, err := versionFromString(startVersionStr)
	if err != nil {
		return nil, err
	}

	// If low and high limits are swapped, fix them
	if highVersion.LessThan(*lowVersion) {
		highVersion, lowVersion = lowVersion, highVersion
	}

	// Special case: same semver as both high and low should just return the
	// single version for fetching but only if that version is actually available
	if highVersion.Equal(*lowVersion) {
		for _, versionStr := range availVersionsStr {
			version, _ := versionFromString(versionStr)
			if version.Equal(*lowVersion) {
				return []string{"v" + lowVersion.String()}, nil
			}
		}

		errorMsg := "v%s is not in available versions (%s)"
		return nil, fmt.Errorf(errorMsg, lowVersion, availVersionsStr)
	}

	versions := []*semver.Version{}
	for _, versionStr := range availVersionsStr {
		// Parse the version from the provided string
		version, err := versionFromString(versionStr)
		if err != nil {
			return nil, err
		}

		// Skip versions higher than highest indicated version.
		if highVersion.LessThan(*version) {
			continue
		}

		// Skip versions lower-or-equal than lowest indicated version.
		if version.LessThan(*lowVersion) || version.Equal(*lowVersion) {
			continue
		}

		versions = append(versions, version)
	}

	// Sort the output since we need to pull data in that order
	semver.Sort(versions)

	// Convert back to strings our version data
	filteredVersionNames := []string{}
	for _, version := range versions {
		filteredVersionNames = append(filteredVersionNames, "v"+(*version).String())
	}

	if len(filteredVersionNames) == 0 {
		errorMsg := "could not find relevant versions for range v%s -> v%s in available versions (%s)"
		return nil, fmt.Errorf(errorMsg, lowVersion, highVersion, availVersionsStr)
	}

	return filteredVersionNames, nil
}
