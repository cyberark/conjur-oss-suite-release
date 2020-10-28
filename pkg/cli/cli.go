package cli

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cyberark/conjur-oss-suite-release/pkg/changelog"
	"github.com/cyberark/conjur-oss-suite-release/pkg/github"
	"github.com/cyberark/conjur-oss-suite-release/pkg/http"
	"github.com/cyberark/conjur-oss-suite-release/pkg/log"
	"github.com/cyberark/conjur-oss-suite-release/pkg/repositories"
	"github.com/cyberark/conjur-oss-suite-release/pkg/template"
	"github.com/cyberark/conjur-oss-suite-release/pkg/version"
)

// Options represents the command line values a user can pass in
type Options struct {
	APIToken           string
	Date               time.Time
	OutputFilename     string
	OutputType         string
	RepositoryFilename string
	ReleasesDir        string
	Version            string
}

type templateInfo struct {
	TemplateName        string
	OutputFilename      string
	VersionInOutputName bool
}

var templates = map[string]templateInfo{
	"changelog": {
		TemplateName:        "CHANGELOG_unified.md.tmpl",
		OutputFilename:      "CHANGELOG_%s.md",
		VersionInOutputName: true,
	},
	"docs-release": {
		TemplateName:        "RELEASE_NOTES_unified.htm.tmpl",
		OutputFilename:      "ConjurSuite_%s.htm",
		VersionInOutputName: true,
	},
	"release": {
		TemplateName:        "RELEASE_NOTES_unified.md.tmpl",
		OutputFilename:      "RELEASE_NOTES_%s.md",
		VersionInOutputName: true,
	},
	"unreleased": {
		TemplateName:        "UNRELEASED_CHANGES_unified.md.tmpl",
		OutputFilename:      "UNRELEASED.md",
		VersionInOutputName: false,
	},
}

const defaultOutputType = "changelog"
const defaultRepositoryFilename = "suite.yml"
const defaultReleasesDir = "releases"
const defaultVersionString = "Unreleased"

// RunParser kicks off the process for writing a new changelog
// 1. Set the output filename if one is not set
// 2. Collect data on each component as specified
// 3. Build a unified changelog with each component
// 4. Write a new changelog based on the appropriate template
func RunParser(options Options) error {
	log.OutLogger.Printf("Parsing linked repositories...")
	repoConfig, err := repositories.NewConfig(options.RepositoryFilename)
	if err != nil {
		return err
	}

	if options.OutputType == "unreleased" {
		// This is an in-place operation
		repoConfig.SelectUnreleased()
	} else if options.ReleasesDir != "" {
		log.OutLogger.Printf("Releases dir: %s", options.ReleasesDir)

		latestReleaseFile, err := version.LatestReleaseInDir(options.ReleasesDir)
		if err != nil {
			return err
		}

		log.OutLogger.Printf("Using %s as previous release for pinning", latestReleaseFile)

		previousReleaseConfig, err := repositories.NewConfig(latestReleaseFile)
		if err != nil {
			return err
		}

		repoConfig.SetBaselineRepoVersions(&previousReleaseConfig)
	}

	log.OutLogger.Printf("Collecting changelogs...")
	httpClient := http.NewClient()

	githubAPIToken := options.APIToken
	if githubAPIToken == "" {
		githubAPIToken = os.Getenv("GITHUB_TOKEN")
	}
	httpClient.AuthToken = githubAPIToken

	suiteCategories, err := github.CollectSuiteCategories(repoConfig, httpClient, options.Version)
	if err != nil {
		return fmt.Errorf("ERROR: %v", err)
	}

	// Combine all changelogs into a single array to generate the unified changelog
	changelogs := []*changelog.VersionChangelog{}
	for _, category := range suiteCategories {
		for _, component := range category.Components {
			if component.Changelogs != nil {
				changelogs = append(changelogs, component.Changelogs...)
			}
		}
	}
	unifiedChangelog := changelog.NewUnifiedChangelog(changelogs...)

	// TODO: Should the date be something defined in yml or the date of tag?
	if options.Date.IsZero() {
		options.Date = time.Now()
	}

	templateData := template.ReleaseSuite{
		// TODO: Suite version should probably be read from some file
		Version:          options.Version,
		Date:             options.Date,
		Description:      repoConfig.Section.Description,
		SuiteCategories:  suiteCategories,
		UnifiedChangelog: unifiedChangelog.String(),
	}

	tmpl := template.New("templates")
	err = tmpl.WriteChangelog(templates[options.OutputType].TemplateName,
		templateData,
		options.OutputFilename)
	if err != nil {
		return err
	}

	log.OutLogger.Printf("Changelog parser completed!")
	return err
}

// HandleInput parses command line values and stores them within an options struct
func (options *Options) HandleInput() error {
	flag.StringVar(&options.RepositoryFilename, "f", defaultRepositoryFilename,
		"Repository YAML file to parse")
	flag.StringVar(&options.ReleasesDir, "r", defaultReleasesDir,
		"Directory of releases (containinng 'suite_<semver>.yml') files. "+
			"Set this to empty string to skip suite version diffing.")
	flag.StringVar(&options.OutputType, "t", defaultOutputType,
		"Output type. Only accepts 'changelog', 'docs-release', 'release', and 'unreleased'.")
	flag.StringVar(&options.OutputFilename, "o", "",
		"Output filename")
	flag.StringVar(&options.Version, "v", defaultVersionString,
		"Version to embed in the changelog")
	flag.StringVar(&options.APIToken, "p", "",
		"GitHub API token. This can also be passed in as the 'GITHUB_TOKEN' environment variable. The flag takes precedence.")
	flag.Parse()

	err := options.setOutputFilename()

	return err
}

func (options *Options) setOutputFilename() error {
	var err error

	tmplInfo, ok := templates[options.OutputType]

	if !ok {
		return fmt.Errorf("%s is not a valid output type", options.OutputType)
	}

	if options.OutputFilename != "" {
		return err
	}

	if tmplInfo.VersionInOutputName {
		options.OutputFilename = fmt.Sprintf(tmplInfo.OutputFilename, options.Version)
		return err
	}

	options.OutputFilename = tmplInfo.OutputFilename
	return err
}
