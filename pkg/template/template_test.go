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

	err = WriteChangelog("testdata/test.tmpl", testObj, outputFile)
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
`

	assert.Equal(t, string(outputFileContent), expectedOutput)
}
