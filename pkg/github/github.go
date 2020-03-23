package github

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cyberark/conjur-oss-suite-release/pkg/changelog"
	"github.com/cyberark/conjur-oss-suite-release/pkg/http"
	"github.com/cyberark/conjur-oss-suite-release/pkg/log"
	"github.com/cyberark/conjur-oss-suite-release/pkg/repositories"
	"github.com/cyberark/conjur-oss-suite-release/pkg/version"
)

// SuiteComponent represents a suite component with all of its changelogs and
// relevant pin data
type SuiteComponent struct {
	CertificationLevel string
	Changelogs         []*changelog.VersionChangelog
	ReleaseName        string
	ReleaseDate        string
	Repo               string
	UpgradeURL         string
	URL                string
}

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

var providerToEndpointPrefix = map[string]string{
	"github": "https://raw.githubusercontent.com",
}

var providerVersionResolutionTemplate = map[string]string{
	"github": "https://api.github.com/repos/%s/releases/latest",
}

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

// FetchChangelog retrieves an existing changelog from a given provider and repository
func fetchChangelog(
	client *http.Client,
	provider string,
	repo string,
	version string,
) (string, error) {

	// `https://raw.githubusercontent.com/cyberark/secretless-broker/master/CHANGELOG.md`
	changelogURL := fmt.Sprintf("%s/%s/%s/CHANGELOG.md", providerToEndpointPrefix[provider], repo, version)
	changelog, err := client.Get(changelogURL)
	if err != nil {
		return "", err
	}

	return string(changelog), nil
}

// CollectComponents retrieves all components specified within a config
func CollectComponents(repoConfig repositories.Config, httpClient *http.Client) (
	[]SuiteComponent,
	error,
) {
	var components []SuiteComponent
	for _, category := range repoConfig.Section.Categories {
		log.OutLogger.Printf("Processing category: %s", category.Name)
		for _, repo := range category.Repos {
			log.OutLogger.Printf("- Processing repo: %s", repo.Name)

			component, err := componentFromRepo(httpClient, repo)
			if err != nil {
				return nil, err
			}

			components = append(components, component)
		}
	}
	return components, nil
}

func componentFromRepo(
	httpClient *http.Client,
	repo repositories.Repository,
) (SuiteComponent, error) {

	component := SuiteComponent{
		Repo:               repo.Name,
		URL:                repo.URL,
		CertificationLevel: repo.CertificationLevel,
		UpgradeURL:         repo.UpgradeURL,
	}

	var changelogs []*changelog.VersionChangelog

	if repo.Version == "" {
		// TODO: This should be somehow transformed from repo url
		version, err := latestVersionToExactVersion(httpClient, "github", repo.Name)
		if err != nil {
			return component, err
		}

		repo.Version = version
	}

	// Repo version is the linked component release version
	component.ReleaseName = repo.Version

	availableVersions, err := GetAvailableReleases(httpClient, repo.Name)
	if err != nil {
		return component, err
	}

	relevantVersions, err := version.GetRelevantVersions(
		availableVersions,
		repo.AfterVersion,
		repo.Version,
	)
	if err != nil {
		return component, err
	}

	log.OutLogger.Printf("  Relevant versions: [%s]", strings.Join(relevantVersions, ", "))

	// TODO: This should be somehow transformed from repo url
	completeChangelog, err := fetchChangelog(httpClient, "github", repo.Name, "master")
	if err != nil {
		return component, err
	}

	// XXX: This still doesn't address releases and how we include that data in yet.
	for _, relevantVersion := range relevantVersions {
		log.OutLogger.Printf("  Extracting changelog data from %s...", relevantVersion)
		versionChangelog, err := extractVersionChangelog(
			repo.Name,
			relevantVersion,
			completeChangelog,
		)
		if err != nil {
			return component, err
		}

		if versionChangelog == nil {
			log.ErrLogger.Printf(
				"  CHANGELOG not found for %s@%s",
				repo.Name,
				relevantVersion,
			)
			continue
		}

		// If this changelog is for the suite release pinned version, save the release
		// date. Since we don't know if there will be `v` prefixes, we compare the
		// strings without them.
		if strings.TrimPrefix(versionChangelog.Version, "v") == strings.TrimPrefix(component.ReleaseName, "v") {
			component.ReleaseDate = versionChangelog.Date
		}

		changelogs = append(changelogs, versionChangelog)
	}

	// Save all relevant component changelogs to the component object
	component.Changelogs = changelogs

	return component, nil
}

func extractVersionChangelog(
	repo string,
	version string,
	log string,
) (*changelog.VersionChangelog, error) {
	versionChangelogs, err := changelog.Parse(repo, log)
	if err != nil {
		return nil, err
	}

	for _, versionChangelog := range versionChangelogs {
		if strings.TrimPrefix(version, "v") == versionChangelog.Version {
			return versionChangelog, nil
		}
	}

	return nil, nil
}

// LatestVersionToExactVersion retrieves the specific version number for a given repo's
// latest version
// https://api.github.com/repos/cyberark/secretless-broker/releases/latest
func latestVersionToExactVersion(client *http.Client, provider string, repo string) (string, error) {
	releaseURL := fmt.Sprintf(providerVersionResolutionTemplate[provider], repo)
	contents, err := client.Get(releaseURL)
	if err != nil {
		return "", err
	}

	var releaseInfo = ReleaseInfo{}
	err = json.Unmarshal(contents, &releaseInfo)
	if err != nil {
		return "", err
	}

	log.OutLogger.Printf("  'latest' resolved as '%s'", releaseInfo.TagName)

	return releaseInfo.TagName, nil
}
