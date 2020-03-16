package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	changelogPkg "github.com/cyberark/conjur-oss-suite-release/pkg/changelog"
	"github.com/cyberark/conjur-oss-suite-release/pkg/github"
	"github.com/cyberark/conjur-oss-suite-release/pkg/http"
	"github.com/cyberark/conjur-oss-suite-release/pkg/log"
	"github.com/cyberark/conjur-oss-suite-release/pkg/repositories"
	"github.com/cyberark/conjur-oss-suite-release/pkg/template"
	"github.com/cyberark/conjur-oss-suite-release/pkg/version"
)

type cliOptions struct {
	APIToken           string
	Date               time.Time
	OutputFilename     string
	OutputType         string
	RepositoryFilename string
	Version            string
}

type templateInfo struct {
	TemplateName string
	OutputName   string
}

var templates = map[string]templateInfo{
	"changelog": {
		TemplateName: "CHANGELOG_unified.md.tmpl",
		OutputName:   "CHANGELOG_%s.md",
	},
	"docs-release": {
		TemplateName: "RELEASE_NOTES_unified.htm.tmpl",
		OutputName:   "ConjurSuite_%s.htm",
	},
	"release": {
		TemplateName: "RELEASE_NOTES_unified.md.tmpl",
		OutputName:   "RELEASE_NOTES_%s.md",
	},
	"unreleased": {
		TemplateName: "UNRELEASED_CHANGES_unified.md.tmpl",
		OutputName:   "UNRELEASED.md",
	},
}

var providerToEndpointPrefix = map[string]string{
	"github": "https://raw.githubusercontent.com",
}

var providerVersionResolutionTemplate = map[string]string{
	"github": "https://api.github.com/repos/%s/releases/latest",
}

const defaultOutputType = "changelog"
const defaultRepositoryFilename = "suite.yml"
const defaultVersionString = "Unreleased"

// https://api.github.com/repos/cyberark/secretless-broker/releases/latest
func latestVersionToExactVersion(client *http.Client, provider string, repo string) (string, error) {
	releaseURL := fmt.Sprintf(providerVersionResolutionTemplate[provider], repo)
	contents, err := client.Get(releaseURL)
	if err != nil {
		return "", err
	}

	var releaseInfo = github.ReleaseInfo{}
	err = json.Unmarshal(contents, &releaseInfo)
	if err != nil {
		return "", err
	}

	log.OutLogger.Printf("  'latest' resolved as '%s'", releaseInfo.TagName)

	return releaseInfo.TagName, nil
}

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

func componentFromRepo(
	httpClient *http.Client,
	repo repositories.Repository,
) (template.SuiteComponent, error) {

	component := template.SuiteComponent{
		Repo:               repo.Name,
		URL:                repo.URL,
		CertificationLevel: repo.CertificationLevel,
		UpgradeURL:         repo.UpgradeURL,
	}

	var changelogs []*changelogPkg.VersionChangelog

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

	availableVersions, err := github.GetAvailableReleases(httpClient, repo.Name)
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
		versionChangelog, err := extractVersionChangeLog(
			repo.Name,
			relevantVersion,
			completeChangelog,
		)
		if err != nil {
			return component, err
		}

		if versionChangelog == nil {
			log.OutLogger.Printf(
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

func collectComponents(repoConfig repositories.Config, httpClient *http.Client) (
	[]template.SuiteComponent,
	error,
) {
	var components []template.SuiteComponent
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

func extractVersionChangeLog(
	repo string,
	version string,
	changelog string,
) (*changelogPkg.VersionChangelog, error) {
	versionChangelogs, err := changelogPkg.Parse(repo, changelog)
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

func runParser(options cliOptions) {
	if _, ok := templates[options.OutputType]; !ok {
		log.ErrLogger.Fatal(fmt.Errorf("%s is not a valid output type", options.OutputType))
		return
	}

	log.OutLogger.Printf("Parsing linked repositories...")
	repoConfig, err := repositories.NewConfig(options.RepositoryFilename)
	if err != nil {
		log.ErrLogger.Fatal(err)
		return
	}

	if options.OutputType == "unreleased" {
		// This is an in-place operation
		repoConfig.SelectUnreleased()
	}

	options.generateOutputFilename()

	log.OutLogger.Printf("Collecting changelogs...")
	httpClient := http.NewClient()

	githubAPIToken := options.APIToken
	if githubAPIToken == "" {
		githubAPIToken = os.Getenv("GITHUB_TOKEN")
	}
	httpClient.AuthToken = githubAPIToken

	components, err := collectComponents(repoConfig, httpClient)
	if err != nil {
		log.ErrLogger.Fatalf("ERROR: %v", err)
		return
	}

	// Combine all changelogs into a single array to generate the unified changelog
	changelogs := []*changelogPkg.VersionChangelog{}
	for _, component := range components {
		changelogs = append(changelogs, component.Changelogs...)
	}
	unifiedChangelog := changelogPkg.NewUnifiedChangelog(changelogs...)

	// TODO: Should the date be something defined in yml or the date of tag?
	if options.Date.IsZero() {
		options.Date = time.Now()
	}

	templateData := template.ReleaseSuite{
		// TODO: Suite version should probably be read from some file
		Version:          options.Version,
		Date:             options.Date,
		Components:       components,
		UnifiedChangelog: unifiedChangelog.String(),
	}

	tmpl := template.New("templates")
	err = tmpl.WriteChangelog(templates[options.OutputType].TemplateName,
		templateData,
		options.OutputFilename)
	if err != nil {
		log.ErrLogger.Fatal(err)
		return
	}

	log.OutLogger.Printf("Changelog parser completed!")
}

// If no filename is provided, we generate one based on the file type
func (options *cliOptions) generateOutputFilename() {
	if options.OutputFilename == "" {
		if options.OutputType == "unreleased" {
			options.OutputFilename = templates[options.OutputType].OutputName
		} else {
			options.OutputFilename = fmt.Sprintf(
				templates[options.OutputType].OutputName,
				options.Version)
		}
	}
}

func main() {
	log.OutLogger.Printf("Starting changelog parser...")

	options := cliOptions{}

	flag.StringVar(&options.RepositoryFilename, "f", defaultRepositoryFilename,
		"Repository YAML file to parse")
	flag.StringVar(&options.OutputType, "t", defaultOutputType,
		"Output type. Only accepts 'changelog', 'docs-release', 'release', and 'unreleased'.")
	flag.StringVar(&options.OutputFilename, "o", "",
		"Output filename")
	flag.StringVar(&options.Version, "v", defaultVersionString,
		"Version to embed in the changelog")
	flag.StringVar(&options.APIToken, "p", "",
		"GitHub API token. This can also be passed in as the 'GITHUB_TOKEN' environment variable. The flag takes precedence.")
	flag.Parse()

	runParser(options)
}
