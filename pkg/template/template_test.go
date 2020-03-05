package template

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteChangelog(t *testing.T) {
	dir, err := ioutil.TempDir("", "changelog_test")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(dir)

	outputFile := filepath.Join(dir, "output.txt")

	testObj := struct {
		StringField string
		ArrayField  []string
	}{
		"somestring",
		[]string{"aaa", "bbb"},
	}

	tmpl := New("testdata")
	err = tmpl.WriteChangelog("test.tmpl", testObj, outputFile)
	if !assert.NoError(t, err) {
		return
	}

	outputFileContent, err := ioutil.ReadFile(outputFile)
	if !assert.NoError(t, err) {
		return
	}

	expectedOutput := `###
somestring

  aaa

  bbb

$$$
abcd
@@@=== PARTIAL START ===
somestring
=== PARTIAL END ===
`

	assert.Equal(t, string(outputFileContent), expectedOutput)
}

func TestWriteChangelogDestinationOpenError(t *testing.T) {
	// Bad path
	outputFile := "doesnotexist/foo"
	testObj := struct{}{}

	tmpl := New("testdata")
	err := tmpl.WriteChangelog("test.tmpl", testObj, outputFile)
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

	assert.EqualError(t, err, "Could not read template 'testdata/doesnotexist'")
}

func TestWriteChangelogTemplateResolutionError(t *testing.T) {
	dir, err := ioutil.TempDir("", "changelog_test")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(dir)

	outputFile := filepath.Join(dir, "output.txt")
	testObj := struct{}{}

	tmpl := New("testdata")
	err = tmpl.WriteChangelog("test.tmpl", testObj, outputFile)
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err,
		"Error running template 'testdata/test.tmpl': "+
			"template: test.tmpl:2:3: executing \"test.tmpl\" at <.StringField>: "+
			"can't evaluate field StringField in type struct {}")
}
