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
	"github.com/cyberark/conjur-oss-suite-release/pkg/github"
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
	conjurReleaseDate1, _ := time.Parse(time.RFC3339, "2020-02-01T11:58:05Z")
	conjurReleaseDate2, _ := time.Parse(time.RFC3339, "2020-01-03T11:58:05Z")
	helmReleaseDate, _ := time.Parse(time.RFC3339, "2020-05-03T11:58:05Z")
	secretlessReleaseDate, _ := time.Parse(time.RFC3339, "2020-01-08T11:58:05Z")

	testData := template.ReleaseSuite{
		Version:          "11.22.33",
		Description:      "A very special suite release.",
		Date:             outputDate,
		UnifiedChangelog: "@@@Unified changelog content@@@",
		SuiteCategories: []github.SuiteCategory{
			github.SuiteCategory{
				CategoryName: "Conjur Core",
				Components: []github.SuiteComponent{
					github.SuiteComponent{
						Repo:                 "cyberark/conjur",
						URL:                  "https://github.com/cyberark/conjur",
						UnreleasedChangesURL: "https://github.com/cyberark/conjur/compare/v1.4.4...HEAD",
						ReleaseName:          "v1.4.4",
						ReleaseDate:          conjurReleaseDate2.Format("2006-01-02"),
						CertificationLevel:   "trusted",
						UpgradeURL:           "https://conjur_upgrade_url",
						Changelogs: []*changelog.VersionChangelog{
							&changelog.VersionChangelog{
								Repo:    "cyberark/conjur",
								Version: "1.3.6",
								// Why are these strings?
								Date: conjurReleaseDate1.Format("2006-01-02"),
								Sections: map[string][]string{
									"Changed": []string{"136Change", "136Change2"},
									"Removed": []string{"136Removal"},
								},
							},
							&changelog.VersionChangelog{
								Repo:    "cyberark/conjur",
								Version: "1.4.4",
								// Why are these strings?
								Date: conjurReleaseDate2.Format("2006-01-02"),
								Sections: map[string][]string{
									"Added":   []string{"144Addition", "144Addition2"},
									"Changed": []string{"144Change", "144Change2"},
									"Fixed":   []string{"144Fix"},
								},
							},
						},
					},
					github.SuiteComponent{
						Repo:                 "cyberark/conjur-oss-helm-chart",
						URL:                  "https://github.com/cyberark/conjur-oss-helm-chart",
						UnreleasedChangesURL: "https://github.com/cyberark/conjur-oss-helm-chart/compare/v1.3.8...HEAD",
						ReleaseName:          "v1.3.8",
						ReleaseDate:          helmReleaseDate.Format("2006-01-02"),
						CertificationLevel:   "trusted",
						Changelogs:           []*changelog.VersionChangelog{},
					},
				},
			},
			github.SuiteCategory{
				CategoryName: "Secrets Delivery",
				Components: []github.SuiteComponent{
					github.SuiteComponent{
						Repo:               "cyberark/secretless-broker",
						URL:                "https://github.com/cyberark/secretless-broker",
						ReleaseName:        "v1.4.2",
						ReleaseDate:        secretlessReleaseDate.Format("2006-01-02"),
						CertificationLevel: "certified",
						Changelogs: []*changelog.VersionChangelog{
							&changelog.VersionChangelog{
								Repo:    "cyberark/secretless-broker",
								Version: "1.4.2",
								Date:    secretlessReleaseDate.Format("2006-01-02"),
								Sections: map[string][]string{
									"Added":   []string{"Broker142Addition", "Broker142Addition With Link [my link](https://github.com/cyberark/conjur/issues/142)"},
									"Changed": []string{"Broker142Change", "Broker142Change With Conjur Docs Link [my link](https://docs.conjur.org/sub-url)"},
									"Removed": []string{"Broker142Removal", "Broker142Removal With CyberArk Docs Link [my link](https://docs.cyberark.com/sub-url)"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range templates {
		t.Run(tt, func(t *testing.T) {
			outputFile := filepath.Join(dir, tt+"_output")

			tmpl := template.New(".")
			err = tmpl.WriteChangelog(tt, testData, outputFile)
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
