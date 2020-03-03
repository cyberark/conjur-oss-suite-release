package template

import (
	"fmt"
	stdlibTemplate "html/template"
	"log"
	"os"
	"time"

	"github.com/cyberark/conjur-oss-suite-release/pkg/changelog"
)

// SuiteComponent represents a suite component with all of its changelogs and
// relevant pin data
type SuiteComponent struct {
	Repo        string
	Changelogs  []*changelog.VersionChangelog
	ReleaseName string
	ReleaseDate string
}

// ReleaseSuite stores all the data needed for generation of templates in the suite
type ReleaseSuite struct {
	Version          string
	Date             time.Time
	Components       []SuiteComponent
	UnifiedChangelog string
}

// WriteChangelog uses a combination of the template path and a data structure
// to create an output file based on that template.
func WriteChangelog(templatePath string,
	templateData interface{},
	outputPath string) error {

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
	tmpl := stdlibTemplate.Must(stdlibTemplate.ParseFiles(templatePath))
	err = tmpl.Execute(outputFile, templateData)
	if err != nil {
		return fmt.Errorf("Error running template '%s': %v", templatePath, err)
	}

	return nil
}
