package version

import (
	"github.com/coreos/go-semver/semver"
)

func versionFromString(versionStr string) (*semver.Version, error) {
	// Strip the 'v' from the beginning, if present
	if versionStr[0] == 'v' {
		versionStr = versionStr[1:]
	}

	return semver.NewVersion(versionStr)
}

// GetRelevantVersions sorts and returns the list of versions from highest
// (included) to the lowest (excluded). The method auto-detects what's the
// lower and what's the higher range bound.
func GetRelevantVersions(availVersionsStr []string,
	startVersionStr string,
	endVersionStr string) ([]string, error) {

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

	// Special case: same semver as both high and low should just
	// return the single version for fetching
	if highVersion.Equal(*lowVersion) {
		return []string{"v" + highVersion.String()}, nil
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

	return filteredVersionNames, nil
}
