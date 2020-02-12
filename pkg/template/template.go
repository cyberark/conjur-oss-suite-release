package template

import (
	"fmt"
	stdlibTemplate "html/template"
	"log"
	"os"
)

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
