package github

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/cyberark/conjur-oss-suite-release/pkg/changelog"
	"github.com/cyberark/conjur-oss-suite-release/pkg/http"
	"github.com/cyberark/conjur-oss-suite-release/pkg/log"
	"github.com/cyberark/conjur-oss-suite-release/pkg/repositories"
	"github.com/cyberark/conjur-oss-suite-release/pkg/version"
)

// SuiteComponent represents a suite component with all of its changelogs and
// relevant pin data
// Note that in #155 we propose moving this into its own package, since it's not
// really relevant to github
type SuiteComponent struct {
	CertificationLevel   string
	Changelogs           []*changelog.VersionChangelog
	ReleaseName          string
	ReleaseDate          string
	Repo                 string
	UnreleasedChangesURL string
	UpgradeURL           string
	URL                  string
}

// SuiteCategory gives the set of SuiteComponents in a category
// Note that in #155 we propose moving this into its own package, since it's not
// really relevant to github
type SuiteCategory struct {
	CategoryName string
	Components   []SuiteComponent
}

// ReleaseInfo is a representation of a v3 GitHub API JSON
// structure denoting a release. We only are interested in
// a small subsection of the field so this list is trimmed
// from the full one that the API returns.
type ReleaseInfo struct {
	Description string `json:"body"`
	Draft       bool   `json:"draft"`
	Name        string `json:"name"`
	TagName     string `json:"tag_name"`
}

// ComparisonInfo is a representation of a v3 GitHub API JSON
// structure denoting a comparison. We only are interested in
// a small subsection of the field so this list is trimmed
// from the full one that the API returns.
type ComparisonInfo struct {
	URL     string `json:"html_url"`
	AheadBy int    `json:"ahead_by"`
}

// e.g. https://api.github.com/repos/cyberark/secretless-broker/releases
const releasesURLTemplate = "https://api.github.com/repos/%s/releases?per_page=100"

// e.g. https://api.github.com/repos/cyberark/secretless-broker/compare/v1.5.2...HEAD
const compareURLTemplate = "https://api.github.com/repos/%s/compare/%s...%s"

// e.g. https://api.github.com/repos/cyberark/secretless-broker/branches/branchName
const branchesURLTemplate = "https://api.github.com/repos/%s/branches/%s"

var providerToEndpointPrefix = map[string]string{
	"github": "https://raw.githubusercontent.com",
}

var providerVersionResolutionTemplate = map[string]string{
	"github": "https://api.github.com/repos/%s/releases/latest",
}

// GetAvailableReleases fetches the JSON at the appropriate URL and returns just
// the available releases as a string array.
func GetAvailableReleases(client http.IClient, repoName string) ([]string, error) {
	return getAvailableReleases(client, fmt.Sprintf(releasesURLTemplate, repoName))
}

func compareRefs(
	client http.IClient,
	repoName string,
	fromRef string,
	toRef string,
) (*ComparisonInfo, error) {
	return comparisonFromURL(
		client,
		fmt.Sprintf(compareURLTemplate, repoName, fromRef, toRef),
	)
}

func comparisonFromURL(
	client http.IClient,
	url string,
) (*ComparisonInfo, error) {
	contents, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	comparison := &ComparisonInfo{}
	err = json.Unmarshal(contents, comparison)
	if err != nil {
		return nil, err
	}

	return comparison, nil
}

func getAvailableReleases(
	client http.IClient,
	releasesURL string,
) ([]string, error) {
	contents, err := client.Get(releasesURL)
	if err != nil {
		return nil, err
	}

	var releases []ReleaseInfo
	err = json.Unmarshal(contents, &releases)
	if err != nil {
		return nil, err
	}

	// Convert ReleaseInfo array to an array of just the version strings
	releaseVersions := make([]string, 0)
	for _, release := range releases {
		versionStr := strings.TrimPrefix(release.Name, "v")
		_, err := semver.NewVersion(versionStr)

		if err != nil {
			// Skip adding versions that don't follow semver
			log.OutLogger.Printf("Skipping version %s", versionStr)
			continue
		}

		releaseVersions = append(releaseVersions, release.Name)
	}

	log.OutLogger.Printf("  Available versions: [%s]", strings.Join(releaseVersions, ", "))

	return releaseVersions, nil
}

// FetchChangelog retrieves an existing changelog from a given provider and repository
func fetchChangelog(
	client http.IClient,
	provider string,
	repo string,
	version string,
) (string, error) {

	// `https://raw.githubusercontent.com/cyberark/secretless-broker/master/CHANGELOG.md`
	changelogURL := fmt.Sprintf("%s/%s/%s/CHANGELOG.md", providerToEndpointPrefix[provider], repo, version)
	changelogBytes, err := client.Get(changelogURL)
	if err != nil {
		return "", err
	}

	return string(changelogBytes), nil
}

// CheckForBranch checks whether a branch with a specific name exists in the repo
func checkForBranch(
	client http.IClient,
	provider string,
	repo string,
	branchName string,
) (bool, error) {

	branchURL := fmt.Sprintf(branchesURLTemplate, repo, url.QueryEscape(branchName))
	_, err := client.Get(branchURL)
	if err != nil {
		if strings.Contains(err.Error(), "Branch not found") {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// CollectSuiteCategories retrieves components for all categories specified
// within a config
func CollectSuiteCategories(repoConfig repositories.Config, httpClient http.IClient, suiteVersion string) (
	[]SuiteCategory,
	error,
) {
	var suiteCategories []SuiteCategory
	for _, category := range repoConfig.Section.Categories {
		log.OutLogger.Printf("Processing category: %s", category.Name)

		var components []SuiteComponent

		for _, repo := range category.Repos {
			log.OutLogger.Printf("- Processing repo: %s", repo.Name)

			component, err := componentFromRepo(httpClient, repo, suiteVersion)
			if err != nil {
				return nil, err
			}

			// No new version between previous and current suite - discard change notes
			if repo.Version == repo.AfterVersion {
				component.Changelogs = nil
			}

			components = append(components, component)
		}

		suiteCategories = append(suiteCategories,
			SuiteCategory{
				CategoryName: category.Name,
				Components:   components,
			})
	}

	return suiteCategories, nil
}

func componentFromRepo(
	httpClient http.IClient,
	repo repositories.Repository,
	suiteVersion string,
) (SuiteComponent, error) {

	component := SuiteComponent{
		Repo:               repo.Name,
		URL:                repo.URL,
		CertificationLevel: repo.CertificationLevel,
		UpgradeURL:         repo.UpgradeURL,
	}

	var changelogs []*changelog.VersionChangelog

	// Repo version is the linked component release version
	component.ReleaseName = repo.Version

	availableVersions, err := GetAvailableReleases(httpClient, repo.Name)
	if err != nil {
		return component, err
	}

	highestVersion, err := version.HighestVersion(availableVersions)
	if err != nil {
		return component, err
	}

	// Get a comparison between the highest version and HEAD
	comparison, err := compareRefs(httpClient, repo.Name, highestVersion, "HEAD")
	if err != nil {
		return component, err
	}

	// Set UnreleasedChangesURL if there are new commits beyond the highest version
	if comparison.AheadBy > 0 {
		component.UnreleasedChangesURL = comparison.URL
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

	// Check if there is a "releases/{suiteVersion}" branch
	// If it exists, use that; if not, use master.
	branch := "master"
	hasReleaseBranch, err := checkForBranch(
		httpClient,
		"github_api",
		repo.Name,
		fmt.Sprintf("release/%s", suiteVersion),
	)
	if err != nil {
		return component, err
	}

	if hasReleaseBranch {
		branch = fmt.Sprintf("release/%s", suiteVersion)
		log.OutLogger.Printf("  Using release branch %s...", branch)
	}

	// Check if the repo has a "main" branch; if so, and there is no matching release
	// branch for this version, use "main" as the default instead
	if !hasReleaseBranch {
		hasMainBranch, err := checkForBranch(
			httpClient,
			"github_api",
			repo.Name,
			"main",
		)

		if err != nil {
			return component, err
		}

		if hasMainBranch {
			branch = "main"

			log.OutLogger.Print("  Using main branch...")
		}
	}

	// TODO: This should be somehow transformed from repo url
	completeChangelog, err := fetchChangelog(httpClient, "github", repo.Name, branch)
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
