package repositories

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type describedObject struct {
	Name        string
	Description string
}

// Repository represents a codified description of a target component
type Repository struct {
	describedObject    `yaml:",inline"`
	URL                string
	CertificationLevel string `yaml:"certification",omitempty`
	Version            string `yaml:omitempty`
	AfterVersion       string `yaml:"after",omitempty`
	UpgradeURL         string `yaml:"upgrade_url",omitempty`
}

type category struct {
	describedObject `yaml:",inline"`
	Repos           []Repository
}

type section struct {
	describedObject `yaml:",inline"`
	Categories      []category
}

// Config is the toplevel object containing the layout of a repositories.yml
// file
type Config struct {
	Section section
}

// NewConfig ingests a YAML file and returns a Config representing the definitions
// in that file.
func NewConfig(filename string) (Config, error) {
	log.Printf("Reading %s...", filename)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("error reading YAML file: %s", err)
	}

	log.Printf("Unmarshaling data...")
	var repoConfig Config
	err = yaml.Unmarshal(yamlFile, &repoConfig)
	if err != nil {
		return Config{}, fmt.Errorf("error unmarshaling YAML file: %s", err)
	}

	return repoConfig, nil
}

// UpdateConfigVersions updates a config's repos' versions in-place, if the repo has a newer
// version in the supplied config. It does this through the following steps:
// 1. Sets the 'afterVersion' to the current 'version' of a repo
// 2. Updates the 'version' to the newer one
func (config *Config) UpdateConfigVersions(configWithNewVersions *Config) {
	// Extract repos and their new versions regardless of category
	newVersions := make(map[string]string)

	for _, category := range configWithNewVersions.Section.Categories {
		for _, repo := range category.Repos {
			newVersions[repo.URL] = repo.Version
		}
	}

	// We use indexes since modifying objects while using them doesn't work in Golang
	// as expected.
	// More info: https://github.com/golang/go/wiki/CommonMistakes#using-reference-to-loop-iterator-variable
	for _, category := range config.Section.Categories {
		for repoIndex, repo := range category.Repos {
			newVersion, present := newVersions[repo.URL]
			if !present {
				continue
			}
			remappedRepo := repo
			remappedRepo.Version = newVersion
			remappedRepo.AfterVersion = repo.Version

			category.Repos[repoIndex] = remappedRepo
		}
	}
}

// SelectUnreleased modifies a Config in-place that will pin all component version
// minimums to the maximums of the input Config as well as unset the maximum, effectively
// enabling us to figure out what a Config for unreleased component versions would
// include.
func (config *Config) SelectUnreleased() {
	for _, category := range config.Section.Categories {
		// We use indexes since modifying objects while using them doesn't work in Golang
		// as expected.
		// More info: https://github.com/golang/go/wiki/CommonMistakes#using-reference-to-loop-iterator-variable
		for repoIndex, repo := range category.Repos {
			remappedRepo := repo
			remappedRepo.Version = ""
			remappedRepo.AfterVersion = repo.Version

			category.Repos[repoIndex] = remappedRepo
		}
	}
}
