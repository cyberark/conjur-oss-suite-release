package github

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cyberark/conjur-oss-suite-release/pkg/log"

	"github.com/cyberark/conjur-oss-suite-release/pkg/http"
)

// ReleaseInfo is a representation of a v3 GitHub API JSON
// structure denitong a release. We only are interested in
// a small subsection of the field so this list is trimmed
// from the full one that the API returns.
type ReleaseInfo struct {
	Description string `json:"body"`
	Draft       bool   `json:"draft"`
	Name        string `json:"name"`
	TagName     string `json:"tag_name"`
}

// eg. https://api.github.com/repos/cyberark/secretless-broker/releases
var releasesURLTemplate = "https://api.github.com/repos/%s/releases"

// GetAvailableReleases fetches the JSON at the appropriate URL and returns just
// the available releases as a string array.
func GetAvailableReleases(client *http.Client, repoName string) ([]string, error) {
	return getAvailableReleases(client, fmt.Sprintf(releasesURLTemplate, repoName))
}

func getAvailableReleases(
	client *http.Client,
	releasesURL string,
) ([]string, error) {

	contents, err := client.Get(releasesURL)
	if err != nil {
		return nil, err
	}

	var releases = []ReleaseInfo{}
	err = json.Unmarshal(contents, &releases)
	if err != nil {
		return nil, err
	}

	// Convert ReleaseInfo array to an array of just the version strings
	releaseVersions := make([]string, len(releases))
	for index, release := range releases {
		releaseVersions[index] = release.TagName
	}

	log.OutLogger.Printf("  Available versions: [%s]", strings.Join(releaseVersions, ", "))

	return releaseVersions, nil
}
