package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestRepoObject(name string) repository {
	return repository{
		describedObject: describedObject{
			Name:        name + " Name",
			Description: name + " Description",
		},
		URL:     name + " URL",
		Version: name + " Version",
	}
}

func TestNewConfig(t *testing.T) {
	expectedRepo1 := newTestRepoObject("Repo1")
	expectedRepo1.AfterVersion = "Repo1 After Version"

	expectedCategories := []category{
		category{
			describedObject: describedObject{
				Name:        "Category1",
				Description: "Category1 Description",
			},
			Repos: []repository{
				expectedRepo1,
				newTestRepoObject("Repo2"),
			},
		},
		category{
			describedObject: describedObject{
				Name:        "Category2",
				Description: "Category2 Description",
			},
			Repos: []repository{
				newTestRepoObject("Repo3"),
			},
		},
	}

	expectedConfig := Config{
		Section: section{
			describedObject: describedObject{
				Name:        "Section Name",
				Description: "Section Description",
			},
			Categories: expectedCategories,
		},
	}

	testPath := "testdata/repositories.yml"

	reposConfig, err := NewConfig(testPath)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, expectedConfig, reposConfig)
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
