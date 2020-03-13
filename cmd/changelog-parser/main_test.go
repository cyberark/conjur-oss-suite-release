package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	// Construct a path to our test repositories yaml
	thisDir, err := os.Getwd()
	if !assert.NoError(t, err) {
		return
	}

	testRepositoriesYml := filepath.Join(thisDir, "testdata", "repositories.yml")

	// We have to run from toplevel dir to be able to use the defaults
	os.Chdir("../..")

	for _, tt := range []string{"changelog", "release"} {
		t.Run(tt, func(t *testing.T) {
			// Create a tempdir to write the out output to
			outputDir, err := ioutil.TempDir("", "main_test")
			if !assert.NoError(t, err) {
				return
			}
			defer os.RemoveAll(outputDir)

			outputFile := filepath.Join(outputDir, tt+"_output.md")
			outputDate, _ := time.Parse(time.RFC3339, "2020-02-19T12:00:00Z")

			// Run the test
			runParser(cliOptions{
				Date:               outputDate,
				OutputFilename:     outputFile,
				OutputType:         tt,
				RepositoryFilename: testRepositoriesYml,
				Version:            "Unreleased",
			})

			log.Print("Verifying test result...")

			outputFileContent, err := ioutil.ReadFile(outputFile)
			if !assert.NoError(t, err) {
				return
			}

			// Tests are expected at "./testdata/expected_<type>_output.md"
			expectedOutputFile := filepath.Join(thisDir, "testdata", "expected_"+tt+"_output.md")
			expectedOutput, err := ioutil.ReadFile(expectedOutputFile)
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, string(expectedOutput), string(outputFileContent))
		})
	}
}

func TestGenerateOutputFilename(t *testing.T) {
	outputDate, _ := time.Parse(time.RFC3339, "2020-02-19T12:00:00Z")
	testCases := []struct {
		description      string
		expectedFilename string
		options          cliOptions
	}{
		{
			description:      "Output type is release, no filename given",
			expectedFilename: "RELEASE_NOTES_Unreleased.md",
			options: cliOptions{
				Date:               outputDate,
				OutputFilename:     "",
				OutputType:         "release",
				RepositoryFilename: "bar",
				Version:            "Unreleased",
			},
		},
		{
			description:      "Output type is changelog, no filename given",
			expectedFilename: "CHANGELOG_Unreleased.md",
			options: cliOptions{
				Date:               outputDate,
				OutputFilename:     "",
				OutputType:         "changelog",
				RepositoryFilename: "bar",
				Version:            "Unreleased",
			},
		},
		{
			description: "Output type is unreleased, no filename given",
			options: cliOptions{
				Date:               outputDate,
				OutputFilename:     "",
				OutputType:         "unreleased",
				RepositoryFilename: "bar",
				Version:            "Unreleased",
			},
			expectedFilename: "UNRELEASED.md",
		},
		{
			description: "Output type is release, filename is given",
			options: cliOptions{
				Date:               outputDate,
				OutputFilename:     "foo.md",
				OutputType:         "release",
				RepositoryFilename: "bar",
				Version:            "Unreleased",
			},
			expectedFilename: "foo.md",
		},
	}

	for _, tc := range testCases {
		t.Run("GenerateOutputFilename: " + tc.description, func(t *testing.T) {
			tc.options.generateOutputFilename()

			assert.EqualValues(t, tc.expectedFilename, tc.options.OutputFilename)
		})
	}
}
