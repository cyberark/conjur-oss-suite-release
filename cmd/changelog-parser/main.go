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

type DescribedObject struct {
	Name        string
	Description string
}

type Repository struct {
	DescribedObject `yaml:",inline"`
	Url             string
	Version         string `yaml:omitempty`
	AfterVersion    string `yaml:"after",omitempty`
}

type Category struct {
	DescribedObject `yaml:",inline"`
	Repos           []Repository
}

type Section struct {
	DescribedObject `yaml:",inline"`
	Categories      []Category
}

type YamlRepoConfig struct {
	Section Section
}

type UnifiedChangelogTemplateData struct {
	Version          string
	Date             time.Time
	UnifiedChangelog string
}

var ProviderToEndpointPrefix = map[string]string{
	"github": "https://raw.githubusercontent.com",
}

var ProviderVersionResolutionTemplate = map[string]string{
	"github": "https://api.github.com/repos/%s/releases/latest",
}

var ProviderReleasesTemplate = map[string]string{
	"github": "https://api.github.com/repos/%s/releases",
}

const DefaultOutputFilename = "CHANGELOG.md"
const DefaultRepositoryFilename = "repositories.yml"
const DefaultVersionString = "Unreleased"

const UnifiedChangelogTemplatePath = "./templates/CHANGELOG_unified.md.tmpl"

func parseLinkedRepositories(filename string) (YamlRepoConfig, error) {
	log.Printf("Reading %s...", filename)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return YamlRepoConfig{},
			fmt.Errorf("Error reading YAML file: %s\n", err)
	}

	log.Printf("Unmarshaling data...")
	var repoConfig YamlRepoConfig
	err = yaml.Unmarshal(yamlFile, &repoConfig)
	if err != nil {
		return YamlRepoConfig{},
			fmt.Errorf("Error reading YAML file: %s\n", err)
	}

	return repoConfig, nil
}

// https://api.github.com/repos/cyberark/secretless-broker/releases/latest
func latestVersionToExactVersion(provider string, repo string) (string, error) {
	releaseUrl := fmt.Sprintf(ProviderVersionResolutionTemplate[provider], repo)
	contents, err := http.Get(releaseUrl)
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
	changelogUrl := fmt.Sprintf("%s/%s/%s/CHANGELOG.md", ProviderToEndpointPrefix[provider], repo, version)
	changelog, err := http.Get(changelogUrl)
	if err != nil {
		return "", err
	}

	return string(changelog), nil
}

// https://api.github.com/repos/cyberark/secretless-broker/releases
func getAvailableReleases(provider string, repo string) ([]string, error) {
	releasesUrl := fmt.Sprintf(ProviderReleasesTemplate[provider], repo)
	contents, err := http.Get(releasesUrl)
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

func collectChangelogs(repoConfig YamlRepoConfig) (
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

	var outputFilename, repositoryFilename, version string
	flag.StringVar(&repositoryFilename, "f", DefaultRepositoryFilename,
		"Repository YAML file to parse")
	flag.StringVar(&outputFilename, "o", DefaultOutputFilename,
		"Output filename")
	flag.StringVar(&version, "v", DefaultVersionString,
		"Version to embed in the changelog")
	flag.Parse()

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

	templateData := UnifiedChangelogTemplateData{
		// TODO: Suite version should probably be read from some file
		// TODO: Should the date be something defined in yml or the date of tag?
		Version:          version,
		Date:             time.Now(),
		UnifiedChangelog: unifiedChangelog.String(),
	}

	err = template.WriteChangelog(UnifiedChangelogTemplatePath,
		templateData,
		outputFilename)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Changelog parser completed!")
}
