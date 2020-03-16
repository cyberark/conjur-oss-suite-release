package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
