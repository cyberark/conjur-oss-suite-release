package changelog

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseChangelog(filename string) ([]*VersionChangelog, error) {
	changelog, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s", filename))
	if err != nil {
		return nil, err
	}

	return Parse("test-repo", string(changelog))
}

func TestParseSimpleChangelog(t *testing.T) {
	changelogs, err := parseChangelog("changelog.simple.md")
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, changelogs[0], &VersionChangelog{
		Repo:    "test-repo",
		Version: "1.5.0",
		Title:   "[1.5.0] 2020-01-29",
		Date:    "2020-01-29",
		Body: `### Added
- add 1
### Changed
- change 1
- change 2`,
		Sections: map[string][]string{
			"Added": {
				"add 1",
			},
			"Changed": {
				"change 1",
				"change 2",
			},
			"_": {
				"add 1",
				"change 1",
				"change 2",
			},
		},
	})
}

func TestComplexChangelog(t *testing.T) {
	changelogs, err := parseChangelog("changelog.complex.md")
	if !assert.NoError(t, err) {
		return
	}


	assert.Equal(t, changelogs[0].Body, `### Changed
- K8s hosts' application identity is extracted from annotations or id. If it is
  defined in annotations it will taken from there and if not, it will be taken
  from the id.
- Another change ABC!@#$%`)

	return
	assert.Equal(t, changelogs[0], &VersionChangelog{
		Repo:    "test-repo",
		Version: "1.4.6",
		Title:   "[1.4.6] - 2020-01-21",
		Date:    "2020-01-21",
		Body: `### Changed
- K8s hosts' application identity is extracted from annotations or id. If it is
defined in annotations it will taken from there and if not, it will be taken
from the id.
- Another change ABC!@#$%`,
		Sections: map[string][]string{
			"Changed": {
				`K8s hosts' application identity is extracted from annotations or id. If it is
defined in annotations it will taken from there and if not, it will be taken
from the id.`,
				"Another change ABC!@#$%",
			},
			"_": {
				`K8s hosts' application identity is extracted from annotations or id. If it is
defined in annotations it will taken from there and if not, it will be taken
from the id.`,
				"Another change ABC!@#$%",
			},
		},
	})
}
