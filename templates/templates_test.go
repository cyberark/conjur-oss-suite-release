package templates

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/conjur-oss-suite-release/pkg/changelog"
	"github.com/cyberark/conjur-oss-suite-release/pkg/template"
)

var templateExt = ".tmpl"

func getTemplatesInDir() ([]string, error) {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var filteredFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == templateExt {
			filteredFiles = append(filteredFiles, file.Name())
		}
	}

	return filteredFiles, nil
}

func TestTemplates(t *testing.T) {
	templates, err := getTemplatesInDir()
	if !assert.NoError(t, err) {
		return
	}

	fmt.Printf("%v", templates)

	dir, err := ioutil.TempDir("", "template_test")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(dir)

	outputDate, _ := time.Parse(time.RFC3339, "2020-02-19T11:58:05Z")
	date1, _ := time.Parse(time.RFC3339, "2020-02-01T11:58:05Z")
	date2, _ := time.Parse(time.RFC3339, "2020-01-03T11:58:05Z")

	testData := template.UnifiedChangelogTemplateData{
		Version:          "11.22.33",
		Date:             outputDate,
		UnifiedChangelog: "@@@Unified changelog content@@@",
		Changelogs: []*changelog.VersionChangelog{
			&changelog.VersionChangelog{
				Repo:    "cyberark/conjur",
				Version: "1.3.6",
				// Why are these strings?
				Date: date1.Format("2006-01-02"),
				Sections: map[string][]string{
					"Changed": []string{"136Change", "136Change2"},
					"Removed": []string{"136Removal"},
				},
			},
			&changelog.VersionChangelog{
				Repo:    "cyberark/conjur",
				Version: "1.4.4",
				// Why are these strings?
				Date: date2.Format("2006-01-02"),
				Sections: map[string][]string{
					"Added":   []string{"144Addition", "144Addition2"},
					"Changed": []string{"144Change", "144Change2"},
					"Fixed":   []string{"144Fix"},
				},
			},
		},
	}

	for _, tt := range templates {
		t.Run(tt, func(t *testing.T) {
			outputFile := filepath.Join(dir, tt+"_output.txt")
			err = template.WriteChangelog(tt, testData, outputFile)
			if !assert.NoError(t, err) {
				return
			}

			outputFileContent, err := ioutil.ReadFile(outputFile)
			if !assert.NoError(t, err) {
				return
			}

			// Tests are expected at "./testdata/<name_without_extension>"
			expectedOutputFile := filepath.Join("testdata", tt[:len(tt)-len(templateExt)])
			expectedOutput, err := ioutil.ReadFile(expectedOutputFile)
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, string(expectedOutput), string(outputFileContent))
		})
	}
}
