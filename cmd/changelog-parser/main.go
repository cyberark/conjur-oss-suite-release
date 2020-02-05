package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
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

	request.Header.Add("Authorization", "token 725ae82aaf466b77aa3e1260cc3f93e64e2d29dc")

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

// patterns
var semverRgx = regexp.MustCompile(`\[?v?([\w\d.-]+\.[\w\d.-]+[a-zA-Z0-9])]?`)
var dateRgx = regexp.MustCompile(`.*[ ](\d\d?\d?\d?[-/.]\d\d?[-/.]\d\d?\d?\d?).*`)

func parseChangelog(changelog string) ([]*VersionChangelog, error) {
	scanner := bufio.NewScanner(strings.NewReader(changelog))

	var versionChangelog *VersionChangelog
	var changeLogs []*VersionChangelog

	for scanner.Scan() {
		line := scanner.Text()

		// skip line if it's a link label
		match, _ := regexp.MatchString(`^\[[^[\]]*\] *?:`, line)
		if match {
			continue
		}

		// new version found!
		match, _ = regexp.MatchString(`^##? ?[^#]`, line)
		if match {
			if versionChangelog != nil && versionChangelog.Title != "" && versionChangelog.Version != "" {
				changeLogs = append(changeLogs, versionChangelog)
			}

			versionChangelog = &VersionChangelog{}

			versionChangelog.Title = string(line)[2:]

			version := semverRgx.FindStringSubmatch(line)
			if len(version) == 0 {
				continue
			}
			versionChangelog.Version = string(version[1])

			date := dateRgx.FindStringSubmatch(versionChangelog.Title)
			if len(date) == 0 {
				continue
			}
			versionChangelog.Date = string(date[1])

			continue
		}

		// accumulate current changelog's body
		if versionChangelog != nil {
			versionChangelog.Body += fmt.Sprintln(line)
		}
	}
	err := scanner.Err()

	return changeLogs, err
}

type VersionChangelog struct {
	Version string
	Title string
	Date string
	Body string
}

func extractVersionChangeLog(
	changelog string,
	version string,
) (string, error) {
	versionChangelogs, err := parseChangelog(changelog)
	if err != nil {
		return "", err
	}

	for _, versionChangelog := range versionChangelogs {
		if strings.TrimPrefix(version, "v") == versionChangelog.Version {
			return versionChangelog.Body, nil
		}
	}

	return "", nil
}
