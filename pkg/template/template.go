package template

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	stdlibTemplate "text/template"
	"time"

	"github.com/cyberark/conjur-oss-suite-release/pkg/changelog"
)

// SuiteComponent represents a suite component with all of its changelogs and
// relevant pin data
type SuiteComponent struct {
	Repo               string
	URL                string
	ReleaseName        string
	ReleaseDate        string
	CertificationLevel string
	Changelogs         []*changelog.VersionChangelog
	UpgradeURL         string
}

// ReleaseSuite stores all the data needed for generation of templates in the suite
type ReleaseSuite struct {
	Version          string
	Date             time.Time
	Components       []SuiteComponent
	UnifiedChangelog string
}

// TemplatesExtension is the extension used for templates that we can use
// to glob-load into the templating engine.
const TemplatesExtension = ".md"

// Engine is a templating generation object that can use partials and helpers
type Engine struct {
	BaseDir string
}

// Define helper methods for templates
var funcMap = stdlibTemplate.FuncMap{
	"toLower": strings.ToLower,
}

// New returns a new templating.Engine based on the specified root
// template directory.
func New(baseDir string) *Engine {
	return &Engine{
		BaseDir: baseDir,
	}
}

// WriteChangelog uses a combination of the template path and a data structure
// to create an output file based on that template.
func (engine *Engine) WriteChangelog(templateName string,
	templateData interface{},
	outputPath string) error {

	templatePath := filepath.Join(engine.BaseDir, templateName)
	partialsGlobPath := filepath.Join(engine.BaseDir, "partials", "*"+TemplatesExtension)

	// Sanity check
	if _, err := os.Stat(templatePath); err != nil {
		return fmt.Errorf("Could not read template '%s'", templatePath)
	}

	// Open the target file
	log.Printf("Opening '%s'...", outputPath)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("Error creating %s: %v", outputPath, err)
	}
	defer outputFile.Close()

	// Generate and write the data to it
	log.Printf("Generating '%s' file from template '%s'...", outputPath, templatePath)

	// Create new parent template with helper functions
	tmpl := stdlibTemplate.New("template").Funcs(funcMap)

	// Load up any needed partials
	tmpl = stdlibTemplate.Must(tmpl.ParseGlob(partialsGlobPath))

	// Load up the requested template
	tmpl = stdlibTemplate.Must(tmpl.ParseFiles(templatePath))

	err = tmpl.ExecuteTemplate(outputFile, templateName, templateData)
	if err != nil {
		return fmt.Errorf("Error running template '%s': %v", templatePath, err)
	}

	return nil
}
