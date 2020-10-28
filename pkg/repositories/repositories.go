package repositories

import (
	"fmt"
	"io/ioutil"

	"github.com/cyberark/conjur-oss-suite-release/pkg/log"

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
	CertificationLevel string `yaml:"certification,omitempty"`
	Version            string `yaml:"version,omitempty"`
	AfterVersion       string `yaml:"after,omitempty"`
	UpgradeURL         string `yaml:"upgrade_url,omitempty"`
}

// Category represents a set of repositories that are logically part of the same
// group
type Category struct {
	describedObject `yaml:",inline"`
	Repos           []Repository
}

// Section is the bulk of the suite configuration, and includes the description
// of this specific release
type Section struct {
	describedObject `yaml:",inline"`
	Categories      []Category
}

// Config is the toplevel object containing the layout of a suite.yml
// file
type Config struct {
	Section Section
}

// NewConfig ingests a YAML file and returns a Config representing the definitions
// in that file.
func NewConfig(filename string) (Config, error) {
	log.OutLogger.Printf("Reading %s...", filename)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("error reading YAML file: %s", err)
	}

	log.OutLogger.Printf("Unmarshaling data...")
	var repoConfig Config
	err = yaml.Unmarshal(yamlFile, &repoConfig)
	if err != nil {
		return Config{}, fmt.Errorf("error unmarshaling YAML file: %s", err)
	}

	return repoConfig, nil
}

// SetBaselineRepoVersions updates the current object with new values for AfterVersion
// field based on the passed in old release config
func (config *Config) SetBaselineRepoVersions(oldConfig *Config) {
	// Extract repos and their old versions regardless of category
	oldVersions := make(map[string]string)
	for _, category := range oldConfig.Section.Categories {
		for _, repo := range category.Repos {
			oldVersions[repo.URL] = repo.Version
		}
	}

	// We use indexes since modifying objects while using them doesn't work in Golang
	// as expected.
	// More info: https://github.com/golang/go/wiki/CommonMistakes#using-reference-to-loop-iterator-variable
	for _, category := range config.Section.Categories {
		for repoIndex, repo := range category.Repos {
			oldVersion, present := oldVersions[repo.URL]
			if !present {
				continue
			}
			remappedRepo := repo
			remappedRepo.AfterVersion = oldVersion

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
