package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	changelogPkg "github.com/cyberark/conjur-oss-suite-release/pkg/changelog"
	"github.com/cyberark/conjur-oss-suite-release/pkg/github"
	"github.com/cyberark/conjur-oss-suite-release/pkg/http"
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

var providerToEndpointPrefix = map[string]string{
	"github": "https://raw.githubusercontent.com",
}

var providerVersionResolutionTemplate = map[string]string{
	"github": "https://api.github.com/repos/%s/releases/latest",
}

var outputTypeTemplates = map[string]string{
	"changelog": "templates/CHANGELOG_unified.md.tmpl",
	"release":   "templates/RELEASE_NOTES_unified.md.tmpl",
}

const defaultOutputFilename = "CHANGELOG.md"
const defaultOutputType = "changelog"
const defaultRepositoryFilename = "repositories.yml"
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

	log.Printf("  'latest' resolved as '%s'", releaseInfo.TagName)

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

func collectChangelogs(repoConfig repositories.Config, httpClient *http.Client) (
	[]*changelogPkg.VersionChangelog,
	error,
) {
	var changelogs []*changelogPkg.VersionChangelog
	for _, category := range repoConfig.Section.Categories {
		log.Printf("Processing category: %s", category.Name)
		for _, repo := range category.Repos {
			log.Printf("- Processing repo: %s", repo.Name)

			if repo.Version == "" {
				// TODO: This should be somehow transformed from repo url
				version, err := latestVersionToExactVersion(httpClient, "github", repo.Name)
				if err != nil {
					return nil, err
				}

				repo.Version = version
			}

			availableVersions, err := github.GetAvailableReleases(httpClient, repo.Name)
			if err != nil {
				return nil, err
			}

			relevantVersions, err := version.GetRelevantVersions(
				availableVersions,
				repo.AfterVersion,
				repo.Version,
			)
			if err != nil {
				return nil, err
			}

			log.Printf("  Relevant versions: [%s]", strings.Join(relevantVersions, ", "))

			// TODO: This should be somehow transformed from repo url
			completeChangelog, err := fetchChangelog(httpClient, "github", repo.Name, "master")
			if err != nil {
				return nil, err
			}

			// XXX: This still doesn't address releases and how we include that data in yet.
			for _, relevantVersion := range relevantVersions {
				log.Printf("  Extracting changelog data from %s...", relevantVersion)
				versionChangelog, err := extractVersionChangeLog(
					repo.Name,
					relevantVersion,
					completeChangelog,
				)
				if err != nil {
					return nil, err
				}

				if versionChangelog == nil {
					log.Printf(
						"  CHANGELOG not found for %s@%s",
						repo.Name,
						relevantVersion,
					)
					continue
				}
				changelogs = append(changelogs, versionChangelog)
			}
		}
	}
	return changelogs, nil
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
	if _, ok := outputTypeTemplates[options.OutputType]; !ok {
		log.Fatal(fmt.Errorf("%s is not a valid output type", options.OutputType))
		return
	}

	log.Printf("Parsing linked repositories...")
	repoConfig, err := repositories.NewConfig(options.RepositoryFilename)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Collecting changelogs...")
	httpClient := http.NewClient()
	httpClient.AuthToken = options.APIToken

	changelogs, err := collectChangelogs(repoConfig, httpClient)
	if err != nil {
		log.Fatal(err)
		return
	}

	// TODO: Same-repo changelogs with different versions should be sorted
	//       in descending order and not ascending one.
	unifiedChangelog := changelogPkg.NewUnifiedChangelog(changelogs...)

	// TODO: Should the date be something defined in yml or the date of tag?
	if options.Date.IsZero() {
		options.Date = time.Now()
	}

	templateData := template.UnifiedChangelogTemplateData{
		// TODO: Suite version should probably be read from some file
		Version:          options.Version,
		Date:             options.Date,
		Changelogs:       changelogs,
		UnifiedChangelog: unifiedChangelog.String(),
	}

	err = template.WriteChangelog(outputTypeTemplates[options.OutputType],
		templateData,
		options.OutputFilename)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Changelog parser completed!")
}

func main() {
	log.Printf("Starting changelog parser...")

	options := cliOptions{}

	flag.StringVar(&options.RepositoryFilename, "f", defaultRepositoryFilename,
		"Repository YAML file to parse")
	flag.StringVar(&options.OutputType, "t", defaultOutputType,
		"Output type. Only accepts 'changelog' and 'release'.")
	flag.StringVar(&options.OutputFilename, "o", defaultOutputFilename,
		"Output filename")
	flag.StringVar(&options.Version, "v", defaultVersionString,
		"Version to embed in the changelog")
	flag.StringVar(&options.APIToken, "p", "",
		"GitHub API token")
	flag.Parse()

	runParser(options)
}
