package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	changelogPkg "github.com/cyberark/conjur-oss-suite-release/pkg/changelog"
	"github.com/cyberark/conjur-oss-suite-release/pkg/github"
	"github.com/cyberark/conjur-oss-suite-release/pkg/http"
	"github.com/cyberark/conjur-oss-suite-release/pkg/template"
	"github.com/cyberark/conjur-oss-suite-release/pkg/version"
)

type describedObject struct {
	Name        string
	Description string
}

type repository struct {
	describedObject `yaml:",inline"`
	URL             string
	Version         string `yaml:omitempty`
	AfterVersion    string `yaml:"after",omitempty`
}

type category struct {
	describedObject `yaml:",inline"`
	Repos           []repository
}

type section struct {
	describedObject `yaml:",inline"`
	Categories      []category
}

type yamlRepoConfig struct {
	Section section
}

var providerToEndpointPrefix = map[string]string{
	"github": "https://raw.githubusercontent.com",
}

var providerVersionResolutionTemplate = map[string]string{
	"github": "https://api.github.com/repos/%s/releases/latest",
}

var providerReleasesTemplate = map[string]string{
	"github": "https://api.github.com/repos/%s/releases",
}

var outputTypeTemplates = map[string]string{
	"changelog": "templates/CHANGELOG_unified.md.tmpl",
	"release":   "templates/RELEASE_NOTES_unified.md.tmpl",
}

const defaultOutputFilename = "CHANGELOG.md"
const defaultOutputType = "changelog"
const defaultRepositoryFilename = "repositories.yml"
const defaultVersionString = "Unreleased"

func parseLinkedRepositories(filename string) (yamlRepoConfig, error) {
	log.Printf("Reading %s...", filename)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return yamlRepoConfig{},
			fmt.Errorf("error reading YAML file: %s", err)
	}

	log.Printf("Unmarshaling data...")
	var repoConfig yamlRepoConfig
	err = yaml.Unmarshal(yamlFile, &repoConfig)
	if err != nil {
		return yamlRepoConfig{},
			fmt.Errorf("error reading YAML file: %s", err)
	}

	return repoConfig, nil
}

// https://api.github.com/repos/cyberark/secretless-broker/releases/latest
func latestVersionToExactVersion(provider string, repo string) (string, error) {
	releaseURL := fmt.Sprintf(providerVersionResolutionTemplate[provider], repo)
	contents, err := http.Get(releaseURL)
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

func fetchChangelog(provider string, repo string, version string) (string, error) {
	// `https://raw.githubusercontent.com/cyberark/secretless-broker/master/CHANGELOG.md`
	changelogURL := fmt.Sprintf("%s/%s/%s/CHANGELOG.md", providerToEndpointPrefix[provider], repo, version)
	changelog, err := http.Get(changelogURL)
	if err != nil {
		return "", err
	}

	return string(changelog), nil
}

// https://api.github.com/repos/cyberark/secretless-broker/releases
func getAvailableReleases(provider string, repo string) ([]string, error) {
	releasesURL := fmt.Sprintf(providerReleasesTemplate[provider], repo)
	contents, err := http.Get(releasesURL)
	if err != nil {
		return nil, err
	}

	var releases = []github.ReleaseInfo{}
	err = json.Unmarshal(contents, &releases)
	if err != nil {
		return nil, err
	}

	// Convert ReleaseInfo array to an array of just the version strings
	releaseVersions := make([]string, len(releases))
	for index, release := range releases {
		releaseVersions[index] = release.TagName
	}

	log.Printf("  Available versions: [%s]", strings.Join(releaseVersions, ", "))

	return releaseVersions, nil
}

func collectChangelogs(repoConfig yamlRepoConfig) (
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
				version, err := latestVersionToExactVersion("github", repo.Name)
				if err != nil {
					return nil, err
				}

				repo.Version = version
			}

			availableVersions, err := getAvailableReleases("github", repo.Name)
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
			completeChangelog, err := fetchChangelog("github", repo.Name, repo.Version)
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

func main() {
	log.Printf("Starting changelog parser...")

	var outputFilename, outputType, repositoryFilename, version string
	flag.StringVar(&repositoryFilename, "f", defaultRepositoryFilename,
		"Repository YAML file to parse")
	flag.StringVar(&outputType, "t", defaultOutputType,
		"Output type. Only accepts 'changelog' and 'release'.")
	flag.StringVar(&outputFilename, "o", defaultOutputFilename,
		"Output filename")
	flag.StringVar(&version, "v", defaultVersionString,
		"Version to embed in the changelog")
	flag.Parse()

	if _, ok := outputTypeTemplates[outputType]; !ok {
		log.Fatal(fmt.Errorf("%s is not a valid output type", outputType))
		return
	}

	log.Printf("Parsing linked repositories...")
	repoConfig, err := parseLinkedRepositories(repositoryFilename)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Collecting changelogs...")
	changelogs, err := collectChangelogs(repoConfig)
	if err != nil {
		log.Fatal(err)
		return
	}

	// TODO: Same-repo changelogs with different versions should be sorted
	//       in descending order and not ascending one.
	unifiedChangelog := changelogPkg.NewUnifiedChangelog(changelogs...)

	templateData := template.UnifiedChangelogTemplateData{
		// TODO: Suite version should probably be read from some file
		// TODO: Should the date be something defined in yml or the date of tag?
		Version:          version,
		Date:             time.Now(),
		Changelogs:       changelogs,
		UnifiedChangelog: unifiedChangelog.String(),
	}

	err = template.WriteChangelog(outputTypeTemplates[outputType],
		templateData,
		outputFilename)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Changelog parser completed!")
}
