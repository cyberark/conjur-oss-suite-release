package template

import (
	"fmt"
	htmlTemplate "html/template"
	"os"
	"path/filepath"
	"strings"
	textTemplate "text/template"
	"time"

	"github.com/cyberark/conjur-oss-suite-release/pkg/changelog"
	"github.com/cyberark/conjur-oss-suite-release/pkg/log"
)

// SuiteComponent represents a suite component with all of its changelogs and
// relevant pin data
type SuiteComponent struct {
	CertificationLevel string
	Changelogs         []*changelog.VersionChangelog
	ReleaseName        string
	ReleaseDate        string
	Repo               string
	UpgradeURL         string
	URL                string
}

// ReleaseSuite stores all the data needed for generation of templates in the suite
type ReleaseSuite struct {
	Version          string
	Date             time.Time
	Components       []SuiteComponent
	UnifiedChangelog string
}

// ComponentReleaseVersion returns the release version, stripped of the v-prefix,
// for a given repo.
func (r ReleaseSuite) ComponentReleaseVersion(repo string) string {
	for _, component := range r.Components {
		if component.Repo == repo {
			return strings.TrimPrefix(component.ReleaseName, "v")
		}
	}

	return ""
}

// MarkdownPartialsExt is the extension used for markdown partials glob matcher
const MarkdownPartialsExt = ".md"

// HTMLPartialsExt is the extension used for HTML partials glob matcher
const HTMLPartialsExt = ".htm"

// HTMLTemplateExt is the extension used for templates that create HTML
// output files
const HTMLTemplateExt = ".htm"

// Engine is a templating generation object that can use partials and helpers
type Engine struct {
	BaseDir     string
	PartialsDir string
}

// Define helper methods for templates
var funcMap = map[string]interface{}{
	"toLower":            strings.ToLower,
	"markdownHeaderLink": markdownHeaderLink,
}

func markdownHeaderLink(repo string) string {
	return strings.Replace(repo, "/", "", -1)
}

// New returns a new templating.Engine based on the specified root
// template directory.
func New(baseDir string) *Engine {
	return &Engine{
		BaseDir:     baseDir,
		PartialsDir: filepath.Join(baseDir, "partials"),
	}
}

func (engine *Engine) renderTemplate(
	templateName string,
	templateData interface{},
	outputFile *os.File,
) error {
	templatePath := filepath.Join(engine.BaseDir, templateName)

	// Sanity check
	if _, err := os.Stat(templatePath); err != nil {
		return fmt.Errorf("Could not read template '%s'", templatePath)
	}

	log.OutLogger.Printf("Generating '%s' file from template '%s'...", outputFile.Name(), templatePath)

	// Sadly while `html/template` and `text/template` share the same API, they
	// use incompatible signatures and parameters so depending on the extension,
	// we have to duplicate our work.
	if filepath.Ext(templatePath) == HTMLTemplateExt {
		// HTML template generation
		tmpl := htmlTemplate.New("template").Funcs(funcMap)
		tmpl = htmlTemplate.Must(
			tmpl.ParseGlob(filepath.Join(engine.PartialsDir, "*"+HTMLPartialsExt)),
		)
		tmpl = htmlTemplate.Must(tmpl.ParseFiles(templatePath))

		return tmpl.ExecuteTemplate(outputFile, templateName, templateData)
	}

	// Markdown template generation
	tmpl := textTemplate.New("template").Funcs(funcMap)
	tmpl = textTemplate.Must(
		tmpl.ParseGlob(filepath.Join(engine.PartialsDir, "*"+MarkdownPartialsExt)),
	)
	tmpl = textTemplate.Must(tmpl.ParseFiles(templatePath))

	return tmpl.ExecuteTemplate(outputFile, templateName, templateData)
}

// WriteChangelog uses a combination of the template path and a data structure
// to create an output file based on that template.
func (engine *Engine) WriteChangelog(
	templateName string,
	templateData interface{},
	outputPath string) error {

	// Open the target file
	log.OutLogger.Printf("Opening '%s'...", outputPath)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("Error creating %s: %v", outputPath, err)
	}
	defer outputFile.Close()

	// Setup templating instance and render output
	err = engine.renderTemplate(templateName, templateData, outputFile)
	if err != nil {
		return fmt.Errorf(
			"Error running template '%s/%s': %v",
			engine.BaseDir,
			templateName,
			err,
		)
	}

	return nil
}
