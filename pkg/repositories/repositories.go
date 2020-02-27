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
