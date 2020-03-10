package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestRepoObject(name string) Repository {
	return Repository{
		describedObject: describedObject{
			Name:        name + " Name",
			Description: name + " Description",
		},
		URL:                name + " URL",
		Version:            name + " Version",
		UpgradeURL:         name + " Upgrade Url",
		CertificationLevel: name + " Certification",
	}
}

func testfileExpectedConfig(expectedRepos ...Repository) Config {

	expectedCategories := []category{
		category{
			describedObject: describedObject{
				Name:        "Category1",
				Description: "Category1 Description",
			},
			Repos: expectedRepos,
		},
		category{
			describedObject: describedObject{
				Name:        "Category2",
				Description: "Category2 Description",
			},
			Repos: []Repository{
				newTestRepoObject("Repo3"),
			},
		},
	}

	return Config{
		Section: section{
			describedObject: describedObject{
				Name:        "Section Name",
				Description: "Section Description",
			},
			Categories: expectedCategories,
		},
	}

}

func TestNewConfig(t *testing.T) {
	testPath := "testdata/repositories.yml"

	reposConfig, err := NewConfig(testPath)
	if !assert.NoError(t, err) {
		return
	}

	expectedRepo1 := newTestRepoObject("Repo1")
	expectedRepo1.AfterVersion = "Repo1 After Version"

	expectedRepo2 := newTestRepoObject("Repo2")
	expectedRepo2.UpgradeURL = ""
	expectedRepo2.CertificationLevel = ""

	expectedRepos := testfileExpectedConfig(expectedRepo1, expectedRepo2)
	assert.Equal(t, expectedRepos, reposConfig)
}

func TestUpdateConfigVersions(t *testing.T) {
	currentTestPath := "testdata/repositories_current.yml"
	newTestPath := "testdata/repositories_new.yml"

	currentConfig, err := NewConfig(currentTestPath)
	if !assert.NoError(t, err) {
		return
	}

	newConfig, err := NewConfig(newTestPath)
	if !assert.NoError(t, err) {
		return
	}

	currentConfig.UpdateConfigVersions(&newConfig)

	expectedRepo1 := newTestRepoObject("Repo1")
	expectedRepo1.AfterVersion = "Repo1 After Version"

	expectedRepo2 := newTestRepoObject("Repo2")
	expectedRepo2.AfterVersion = "Repo2 Version"
	expectedRepo2.CertificationLevel = ""
	expectedRepo2.Version = "Repo2 New Version"
	expectedRepo2.UpgradeURL = ""

	expectedRepos := testfileExpectedConfig(expectedRepo1, expectedRepo2)

	assert.Equal(t, expectedRepos, currentConfig)
}

func TestNewConfigReadFileProblems(t *testing.T) {
	_, err := NewConfig("doesnotexist")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"error reading YAML file: open doesnotexist: no such file or directory",
	)
}

func TestNewConfigUnmarshalingProblem(t *testing.T) {
	_, err := NewConfig("./testdata/bad_repositories.yml")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"error unmarshaling YAML file: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `foobar` into repositories.Config",
	)
}

func TestSelectUnreleased(t *testing.T) {
	expectedRepo1 := newTestRepoObject("Repo1")
	expectedRepo1.AfterVersion = "Repo1 After Version"

	expectedRepo2 := newTestRepoObject("Repo2")
	expectedRepo2.UpgradeURL = ""
	expectedRepo2.CertificationLevel = ""

	expectedConfig := testfileExpectedConfig(expectedRepo1, expectedRepo2)
	expectedConfig.Section.Categories[0].Repos[0].Version = ""
	expectedConfig.Section.Categories[0].Repos[0].AfterVersion = "Repo1 Version"
	expectedConfig.Section.Categories[0].Repos[1].Version = ""
	expectedConfig.Section.Categories[0].Repos[1].AfterVersion = "Repo2 Version"
	expectedConfig.Section.Categories[1].Repos[0].Version = ""
	expectedConfig.Section.Categories[1].Repos[0].AfterVersion = "Repo3 Version"

	config, err := NewConfig("./testdata/repositories.yml")
	if !assert.NoError(t, err) {
		return
	}

	config.SelectUnreleased()

	assert.Equal(t, expectedConfig, config)
}
