package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunParser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	// Construct a path to our test repositories yaml
	thisDir, err := os.Getwd()
	if !assert.NoError(t, err) {
		return
	}

	testRepositoriesYml := filepath.Join(thisDir, "testdata", "suite.yml")

	// We have to run from toplevel dir to be able to use the defaults
	os.Chdir("../..")
	defer func() {
		os.Chdir(thisDir)
	}()

	for _, tt := range []string{"changelog", "docs-release", "release"} {
		t.Run(tt, func(t *testing.T) {
			// Create a tempdir to write the out output to
			outputDir, err := ioutil.TempDir("", "main_test")
			if !assert.NoError(t, err) {
				return
			}
			defer os.RemoveAll(outputDir)

			outputFile := filepath.Join(outputDir, tt+"_output.txt")
			outputDate, _ := time.Parse(time.RFC3339, "2020-02-19T12:00:00Z")

			// Run the test
			err = RunParser(Options{
				Date:               outputDate,
				OutputFilename:     outputFile,
				OutputType:         tt,
				RepositoryFilename: testRepositoriesYml,
				Version:            "Unreleased",
			})
			if !assert.NoError(t, err) {
				return
			}

			outputFileContent, err := ioutil.ReadFile(outputFile)
			if !assert.NoError(t, err) {
				return
			}

			// Tests are expected at "./testdata/expected_<type>_output.txt"
			expectedOutputFile := filepath.Join(thisDir, "testdata", "expected_"+tt+"_output.txt")
			expectedOutput, err := ioutil.ReadFile(expectedOutputFile)
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, string(expectedOutput), string(outputFileContent))
		})
	}
}

func TestRunParserWithReleaseDiffing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	// Construct a path to our test repositories yaml
	thisDir, err := os.Getwd()
	if !assert.NoError(t, err) {
		return
	}

	testRepositoriesYml := filepath.Join(thisDir, "testdata", "new_release_suite.yml")

	// We have to run from toplevel dir to be able to use the defaults
	os.Chdir("../..")
	defer func() {
		os.Chdir(thisDir)
	}()

	// Create a tempdir to write the out output to
	outputDir, err := ioutil.TempDir("", "suite_diff_test")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(outputDir)

	for _, tt := range []string{"changelog", "docs-release", "release"} {
		t.Run(tt, func(t *testing.T) {
			outputFile := filepath.Join(outputDir, tt+"output.txt")
			outputDate, _ := time.Parse(time.RFC3339, "2020-02-19T12:00:00Z")

			// Run the test
			err = RunParser(Options{
				Date:               outputDate,
				OutputFilename:     outputFile,
				OutputType:         tt,
				RepositoryFilename: testRepositoriesYml,
				ReleasesDir:        filepath.Join(thisDir, "testdata", "mock_releases"),
				Version:            "Unreleased",
			})
			if !assert.NoError(t, err) {
				return
			}

			outputFileContent, err := ioutil.ReadFile(outputFile)
			if !assert.NoError(t, err) {
				return
			}

			expectedOutputFile := filepath.Join(thisDir, "testdata", "expected_"+tt+"_diff_output.txt")
			expectedOutput, err := ioutil.ReadFile(expectedOutputFile)
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, string(expectedOutput), string(outputFileContent))
		})
	}
}

func TestSettingOutputFilename(t *testing.T) {
	testCases := []struct {
		outputType       string
		version          string
		outputFilename   string
		expectedFilename string
	}{
		{
			outputType:       "changelog",
			version:          "mychangelogversion",
			outputFilename:   "",
			expectedFilename: "CHANGELOG_mychangelogversion.md",
		},
		{
			outputType:       "changelog",
			version:          "foo",
			outputFilename:   "outputname1",
			expectedFilename: "outputname1",
		},
		{
			outputType:       "docs-release",
			version:          "mydocsreleaseversion",
			outputFilename:   "",
			expectedFilename: "ConjurSuite_mydocsreleaseversion.htm",
		},
		{
			outputType:       "docs-release",
			version:          "bar",
			outputFilename:   "outputname2",
			expectedFilename: "outputname2",
		},
		{
			outputType:       "release",
			version:          "myreleaseversion",
			outputFilename:   "",
			expectedFilename: "RELEASE_NOTES_myreleaseversion.md",
		},
		{
			outputType:       "release",
			version:          "baz",
			outputFilename:   "outputname3",
			expectedFilename: "outputname3",
		},
		{
			outputType:       "unreleased",
			version:          "notused",
			outputFilename:   "",
			expectedFilename: "UNRELEASED.md",
		},
		{
			outputType:       "unreleased",
			version:          "notused",
			outputFilename:   "outputname4",
			expectedFilename: "outputname4",
		},
	}

	for _, tc := range testCases {
		options := Options{
			OutputType:     tc.outputType,
			OutputFilename: tc.outputFilename,
			Version:        tc.version,
		}

		testName := fmt.Sprintf(
			"GenerateOutputFilename: %s/'%s'",
			tc.outputType,
			tc.expectedFilename,
		)

		t.Run(testName, func(t *testing.T) {
			options.setOutputFilename()

			assert.EqualValues(t, tc.expectedFilename, options.OutputFilename)
		})
	}
}
