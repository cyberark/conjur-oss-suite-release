package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCombinedChangelog(t *testing.T) {
	expected := CombinedChangelog{
		"ADDED": {
			"`x-repo@x-version`: add 1",
			"`y-repo@y-version`: add 2",
		},
		"CHANGED": {
			"`x-repo@x-version`: change 1",
			"`x-repo@x-version`: change 2",
			"`y-repo@y-version`: change 3",
			"`y-repo@y-version`: change 4",
		},
	}
	actual := NewCombinedChangelog(
		&VersionChangelog{
			Repo: "x-repo",
			Version: "x-version",
			Sections: map[string][]string{
				"ADded": {"add 1"},
				"Changed": {"change 1", "change 2"},
				"_": {"add 1", "change 1", "change 2"},
			},
		},
		&VersionChangelog{
			Repo: "y-repo",
			Version: "y-version",
			Sections: map[string][]string{
				"Added": {"add 2"},
				"changed": {"change 3", "change 4"},
				"_": {"add 2", "change 3", "change 4"},
			},
		},
	)

	assert.EqualValues(t, expected, actual)
}

func TestCombinedChangelog_String(t *testing.T) {
	expected :=`### ADDED
- ` + "`x-repo@x-version`" + `: add 1
- ` + "`y-repo@y-version`" + `: add 2

### CHANGED
- ` + "`x-repo@x-version`" + `: change 1
- ` + "`x-repo@x-version`" + `: change 2
- ` + "`y-repo@y-version`" + `: change 3
- ` + "`y-repo@y-version`" + `: change 4

`
	actual := CombinedChangelog{
		"ADDED": {
			"`x-repo@x-version`: add 1",
			"`y-repo@y-version`: add 2",
		},
		"CHANGED": {
			"`x-repo@x-version`: change 1",
			"`x-repo@x-version`: change 2",
			"`y-repo@y-version`: change 3",
			"`y-repo@y-version`: change 4",
		},
	}.String()


	assert.EqualValues(t, expected, actual)
}
