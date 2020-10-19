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
		Date:    "2020-01-29",
		Sections: map[string][]string{
			"Added": {
				"add 1",
				"cyberark/conjur@1.4.4: Bumped toolset from 3.12.0 to 3.12.2",
			},
			"Changed": {
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

	assert.Equal(t, changelogs[0], &VersionChangelog{
		Repo:    "test-repo",
		Version: "1.4.6",
		Date:    "2020-01-21",
		Sections: map[string][]string{
			"Changed": {
				`K8s hosts' resource restrictions is extracted from annotations or id. If it is
defined in annotations it will taken from there and if not, it will be taken
from the id.`,
				"Another change ABC!@#$%",
			},
		},
	})
}
