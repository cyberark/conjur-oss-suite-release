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
			outputFile := filepath.Join(dir, tt.templateName+"_output.md")

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

	outputFile := filepath.Join(dir, "output.md")
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
			outputFile := filepath.Join(dir, tt.templateName+"_output.md")
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

func TestMarkdownHyperlinksToHTMLHyperlinks(t *testing.T) {
	testData := []struct {
		description    string
		inputString    string
		expectedString string
	}{
		{
			description:    "no url in input",
			inputString:    "foo",
			expectedString: "foo",
		},
		{
			description:    "input includes words in perentheses",
			inputString:    "foo (bar)",
			expectedString: "foo (bar)",
		},
		{
			description:    `input contains a single url`,
			inputString:    `foo [bar](baz)`,
			expectedString: `foo <a href="baz" target="_blank">bar</a>`,
		},
		{
			description:    `input contains a single conjur docs url`,
			inputString:    `foo [bar](https://docs.conjur.org/baz)`,
			expectedString: `foo <a href="https://docs.conjur.org/baz">bar</a>`,
		},
		{
			description:    `input contains a single cyberark docs url`,
			inputString:    `foo [bar](https://docs.cyberark.com/baz)`,
			expectedString: `foo <a href="https://docs.cyberark.com/baz">bar</a>`,
		},
		{
			description:    `input contains multiple urls`,
			inputString:    `foo [bar](baz) & [jack](box)`,
			expectedString: `foo <a href="baz" target="_blank">bar</a> & <a href="box" target="_blank">jack</a>`,
		},
	}

	for _, td := range testData {
		t.Run(td.description, func(t *testing.T) {
			actualString := markdownHyperlinksToHTMLHyperlinks(td.inputString)

			assert.EqualValues(t, td.expectedString, actualString)
		})
	}
}
