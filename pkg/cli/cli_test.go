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
		tmplInfo := templates[tc.outputType]
		options := Options{
			OutputFilename: tc.outputFilename,
			Version:        tc.version,
		}

		testName := fmt.Sprintf(
			"GenerateOutputFilename: %s/'%s'",
			tc.outputType,
			tc.expectedFilename,
		)

		t.Run(testName, func(t *testing.T) {
			options.setOutputFilename(tmplInfo)

			assert.EqualValues(t, tc.expectedFilename, options.OutputFilename)
		})
	}
}

func TestGetTemplateInfo(t *testing.T) {
	testData := []struct {
		description          string
		expectedTemplateInfo templateInfo
		outputType           string
		expectError          bool
	}{
		{
			description: "Valid outputType",
			expectedTemplateInfo: templateInfo{
				TemplateName:        "CHANGELOG_unified.md.tmpl",
				OutputFilename:      "CHANGELOG_%s.md",
				VersionInOutputName: true,
			},
			outputType:  "changelog",
			expectError: false,
		},
		{
			description:          "Invalid outputType",
			expectedTemplateInfo: templateInfo{},
			expectError:          true,
			outputType:           "foo",
		},
	}

	for _, tc := range testData {
		t.Run(tc.description, func(t *testing.T) {
			templateInfo, err := getTemplateInfo(tc.outputType)

			if tc.expectError {
				assert.Error(t, err)
			}

			assert.EqualValues(t, templateInfo, tc.expectedTemplateInfo)
		})
	}
}
