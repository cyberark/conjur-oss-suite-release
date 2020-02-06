package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"

	changelogPkg "github.com/cyberark/conjur-oss-suite-release/pkg/changelog"
)

type DescribedObject struct {
	Name        string
	Description string
}

type Repository struct {
	DescribedObject `yaml:",inline"`
	Url             string
	Version         string `yaml:omitempty`
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

type ReleaseInfo struct {
	TagName string `json:"tag_name"`
}

var ProviderToEndpointPrefix = map[string]string{
	"github": "https://raw.githubusercontent.com",
}

var ProviderVersionResolutionTemplate = map[string]string{
	"github": "https://api.github.com/repos/%s/releases/latest",
}

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
	client := &http.Client{}

	releaseUrl := fmt.Sprintf(ProviderVersionResolutionTemplate[provider], repo)
	request, err := http.NewRequest("GET", releaseUrl, nil)
	if err != nil {
		return "", err
	}

	log.Printf("  Fetching %s release info...", releaseUrl)
	response, err := client.Do(request)

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if response.StatusCode >= 300 {
		return "", fmt.Errorf("code %d: %s: %s", response.StatusCode, releaseUrl, contents)
	}

	var releaseInfo = ReleaseInfo{}
	err = json.Unmarshal(contents, &releaseInfo)
	if err != nil {
		return "", err
	}

	log.Printf("  'latest' -> '%s'", releaseInfo.TagName)

	return releaseInfo.TagName, nil
}

func fetchChangelog(provider string, repo string, version string) (string, error) {
	client := &http.Client{}

	// `https://raw.githubusercontent.com/cyberark/secretless-broker/master/CHANGELOG.md`
	changelogUrl := fmt.Sprintf("%s/%s/%s/CHANGELOG.md", ProviderToEndpointPrefix[provider], repo, version)
	request, err := http.NewRequest("GET", changelogUrl, nil)
	if err != nil {
		return "", err
	}

	log.Printf("  Fetching %s...", changelogUrl)
	response, err := client.Do(request)

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if response.StatusCode >= 300 {
		return "", fmt.Errorf("code %d: %s: %s", response.StatusCode, changelogUrl, contents)
	}

	return string(contents), nil
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
			// TODO: This should be somehow transformed from repo url
			completeChangelog, err := fetchChangelog("github", repo.Name, repo.Version)
			if err != nil {
				return nil, err
			}

			versionChangelog, err := extractVersionChangeLog(
				repo.Name,
				repo.Version,
				completeChangelog,
			)
			if err != nil {
				return nil, err
			}

			changelogs = append(changelogs, versionChangelog)
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

	var outputFilename, repositoryFilename string
	flag.StringVar(&repositoryFilename, "f", "repositories.yml", "Repository YAML file to parse.")
	flag.StringVar(&outputFilename, "o", "CHANGELOG.md", "Output file. Defaults to CHANGELOG.md")
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

	combinedChangelog := changelogPkg.NewCombinedChangelog(changelogs...)

	log.Printf("Writing changelog to %s...", outputFilename)
	err = ioutil.WriteFile(outputFilename, []byte(combinedChangelog.String()), 0644)
	if err != nil {
		panic(err)
	}

	log.Printf("Changelog parser completed!")
}
