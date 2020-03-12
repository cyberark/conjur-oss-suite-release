package template

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const complexString = "!@#$%^&*()<>"
const escapedComplexString = "!@#$%^&amp;*()&lt;&gt;"

var templateTypeTests = []struct {
	templateName         string
	expectedStringOutput string
}{
	{"test.md", complexString},
	{"test.htm", escapedComplexString},
}

func TestWriteChangelog(t *testing.T) {
	dir, err := ioutil.TempDir("", "changelog_test")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(dir)

	for _, tt := range templateTypeTests {
		t.Run(tt.templateName, func(t *testing.T) {
			outputFile := filepath.Join(dir, tt.templateName+"_output.txt")

			testObj := struct {
				StringField string
				ArrayField  []string
			}{
				complexString,
				[]string{"aaa", "bbb"},
			}

			tmpl := New("testdata")
			err = tmpl.WriteChangelog(tt.templateName, testObj, outputFile)
			if !assert.NoError(t, err) {
				return
			}

			outputFileContent, err := ioutil.ReadFile(outputFile)
			if !assert.NoError(t, err) {
				return
			}

			expectedOutput := "###\n" +
				tt.expectedStringOutput +
				"\n\n  aaa" +
				"\n\n  bbb" +
				"\n\n$$$" +
				"\nabcd" +
				"\n@@@" +
				"=== PARTIAL START ===" +
				"\n" + tt.expectedStringOutput +
				"\n=== PARTIAL END ===\n"

			assert.Equal(t, expectedOutput, string(outputFileContent))
		})
	}
}

func TestWriteChangelogDestinationOpenError(t *testing.T) {
	// Bad path
	outputFile := "doesnotexist/foo"
	testObj := struct{}{}

	tmpl := New("testdata")
	err := tmpl.WriteChangelog("test.md", testObj, outputFile)
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err,
		"Error creating doesnotexist/foo: "+
			"open doesnotexist/foo: no such file or directory")
}

func TestWriteChangelogTemplateOpenError(t *testing.T) {
	dir, err := ioutil.TempDir("", "changelog_test")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(dir)

	outputFile := filepath.Join(dir, "output.txt")
	testObj := struct{}{}

	tmpl := New("testdata")
	err = tmpl.WriteChangelog("doesnotexist", testObj, outputFile)
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(
		t,
		err,
		"Error running template 'testdata/doesnotexist': "+
			"Could not read template 'testdata/doesnotexist'",
	)
}

func TestWriteChangelogTemplateResolutionError(t *testing.T) {
	dir, err := ioutil.TempDir("", "changelog_test")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(dir)

	for _, tt := range templateTypeTests {
		t.Run(tt.templateName, func(t *testing.T) {
			outputFile := filepath.Join(dir, tt.templateName+"_output.txt")
			testObj := struct{}{}

			tmpl := New("testdata")
			err = tmpl.WriteChangelog(tt.templateName, testObj, outputFile)
			if !assert.Error(t, err) {
				return
			}

			assert.EqualError(t, err,
				"Error running template 'testdata/"+tt.templateName+"': "+
					"template: "+tt.templateName+":2:3: "+
					"executing \""+tt.templateName+"\" at <.StringField>: "+
					"can't evaluate field StringField in type struct {}")
		})
	}
}
