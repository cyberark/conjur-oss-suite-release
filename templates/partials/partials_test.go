package partials

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

const PartialsExtension = ".md"

func TestCertificationBadge(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "cert_level_partial_test")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(dir)

	tests := []string{
		"trusted",
		"certified",
		"community",
		"unknown",
	}

	testfilePrefix := "certification_badge"

	for _, tt := range tests {
		testData := struct {
			CertificationLevel string
			URL                string
		}{
			CertificationLevel: tt,
			URL:                "http://repo-url",
		}

		t.Run(tt, func(t *testing.T) {
			funcMap := template.FuncMap{
				"toLower": strings.ToLower,
			}

			var actualOutput bytes.Buffer
			tmpl := template.Must(
				template.New("test").Funcs(funcMap).ParseFiles(testfilePrefix + PartialsExtension),
			)

			err := tmpl.ExecuteTemplate(
				&actualOutput,
				testfilePrefix+PartialsExtension,
				testData,
			)
			if !assert.NoError(t, err) {
				return
			}

			expectedOutputFilename := testfilePrefix + "_" + tt + PartialsExtension
			expectedOutputFile := filepath.Join("testdata", expectedOutputFilename)
			expectedOutput, err := ioutil.ReadFile(expectedOutputFile)
			if !assert.NoError(t, err) {
				return
			}

			// Newline is auto-added by editors to our comparison files so we need
			// to add it to our partials too
			assert.Equal(t, string(expectedOutput), actualOutput.String()+"\n")
		})
	}
}
