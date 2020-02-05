package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
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

func collectChangelogs(repoConfig YamlRepoConfig) (map[string]string, error) {
	changelogs := map[string]string{}
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
			changelog, err := fetchChangelog("github", repo.Name, repo.Version)
			if err != nil {
				return nil, err
			}

			changelog, err = extractVersionChangeLog(changelog, repo.Version)
			if err != nil {
				return nil, err
			}

			changelogs[repo.Url] = changelog
			log.Printf("  Changelog size: %d", len(changelog))
		}
	}
	return changelogs, nil
}

func main() {
	log.Printf("Starting changelog parser...")

	var filename string
	flag.StringVar(&filename, "f", "", "Repository YAML file to parse.")
	flag.Parse()

	if filename == "" {
		log.Fatal("Please provide repository YAML file by using -f option")
		return
	}

	log.Printf("Parsing linked repositories...")
	repoConfig, err := parseLinkedRepositories(filename)
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

	log.Printf("Changelogs")
	var res string
	for url, changelog := range changelogs {
		if strings.TrimSpace(changelog) == "" {
			continue
		}
		res += "## " + url + "\n"
		res += changelog + "\n"
	}
	err = ioutil.WriteFile("tmpCHANGELOG.md", []byte(res), 0644)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Changelog parser completed!")
}

func extractVersionChangeLog(
	changelog string,
	version string,
) (string, error) {
	out, err := exec.Command(
		"node",
		"index.js",
		changelog,
		version,
	).Output()

	if err != nil {
		if exitErrr, ok := err.(*exec.ExitError); ok {
			return "", errors.New(string(exitErrr.Stderr))
		}

		return "", err
	}

	return string(out), nil
}