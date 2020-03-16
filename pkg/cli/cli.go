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
)

// Options represents the command line values a user can pass in
type Options struct {
	APIToken           string
	Date               time.Time
	OutputFilename     string
	OutputType         string
	RepositoryFilename string
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
const defaultVersionString = "Unreleased"

// RunParser kicks off the process for writing a new changelog
// 1. Set the output filename if one is not set
// 2. Collect data on each component as specified
// 3. Build a unified changelog with each component
// 4. Write a new changelog based on the appropriate template
func RunParser(options Options) error {
	templateInfo, err := getTemplateInfo(options.OutputType)
	if err != nil {
		return err
	}
	options.setOutputFilename(templateInfo)

	log.OutLogger.Printf("Parsing linked repositories...")
	repoConfig, err := repositories.NewConfig(options.RepositoryFilename)
	if err != nil {
		return err
	}

	if options.OutputType == "unreleased" {
		// This is an in-place operation
		repoConfig.SelectUnreleased()
	}

	log.OutLogger.Printf("Collecting changelogs...")
	httpClient := http.NewClient()

	githubAPIToken := options.APIToken
	if githubAPIToken == "" {
		githubAPIToken = os.Getenv("GITHUB_TOKEN")
	}
	httpClient.AuthToken = githubAPIToken

	components, err := github.CollectComponents(repoConfig, httpClient)
	if err != nil {
		return fmt.Errorf("ERROR: %v", err)
	}

	// Combine all changelogs into a single array to generate the unified changelog
	changelogs := []*changelog.VersionChangelog{}
	for _, component := range components {
		changelogs = append(changelogs, component.Changelogs...)
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
		Components:       components,
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
func (options *Options) HandleInput() {
	flag.StringVar(&options.RepositoryFilename, "f", defaultRepositoryFilename,
		"Repository YAML file to parse")
	flag.StringVar(&options.OutputType, "t", defaultOutputType,
		"Output type. Only accepts 'changelog', 'docs-release', 'release', and 'unreleased'.")
	flag.StringVar(&options.OutputFilename, "o", "",
		"Output filename")
	flag.StringVar(&options.Version, "v", defaultVersionString,
		"Version to embed in the changelog")
	flag.StringVar(&options.APIToken, "p", "",
		"GitHub API token. This can also be passed in as the 'GITHUB_TOKEN' environment variable. The flag takes precedence.")
	flag.Parse()
}

func (options *Options) setOutputFilename(tmplInfo templateInfo) {
	if options.OutputFilename != "" {
		return
	}

	if tmplInfo.VersionInOutputName {
		options.OutputFilename = fmt.Sprintf(tmplInfo.OutputFilename, options.Version)
		return
	}

	options.OutputFilename = tmplInfo.OutputFilename
}

// GetTemplateInfo checks that a given outputType exists within a predefined set of
// template output types, and returns the corresponding info
func getTemplateInfo(outputType string) (templateInfo, error) {
	var templateInfo templateInfo
	var ok bool

	if templateInfo, ok = templates[outputType]; !ok {
		return templateInfo, fmt.Errorf("%s is not a valid output type", outputType)
	}

	return templateInfo, nil
}
