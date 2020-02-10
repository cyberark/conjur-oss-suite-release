package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUnifiedChangelog(t *testing.T) {
	expected := UnifiedChangelog{
		"Added": {
			"`x-repo@x-version`: add 1",
			"`y-repo@y-version`: add 2",
		},
		"Changed": {
			"`x-repo@x-version`: change 1",
			"`x-repo@x-version`: change 2",
			"`y-repo@y-version`: change 3",
			"`y-repo@y-version`: change 4",
		},
	}
	actual := NewUnifiedChangelog(
		&VersionChangelog{
			Repo:    "x-repo",
			Version: "x-version",
			Sections: map[string][]string{
				"ADded": {
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
		},
		&VersionChangelog{
			Repo:    "y-repo",
			Version: "y-version",
			Sections: map[string][]string{
				"Added": {
					"add 2",
				},
				"changed": {
					"change 3",
					"change 4",
				},
				"_": {
					"add 2",
					"change 3",
					"change 4",
				},
			},
		},
	)

	assert.EqualValues(t, expected, actual)
}

func TestUnifiedChangelog_String(t *testing.T) {
	expected := `### Added
- ` + "`x-repo@x-version`" + `: add 1
- ` + "`y-repo@y-version`" + `: add 2

### Changed
- ` + "`x-repo@x-version`" + `: change 1
- ` + "`x-repo@x-version`" + `: change 2
- ` + "`y-repo@y-version`" + `: change 3
- ` + "`y-repo@y-version`" + `: change 4

`
	actual := UnifiedChangelog{
		"Added": {
			"`x-repo@x-version`: add 1",
			"`y-repo@y-version`: add 2",
		},
		"Changed": {
			"`x-repo@x-version`: change 1",
			"`x-repo@x-version`: change 2",
			"`y-repo@y-version`: change 3",
			"`y-repo@y-version`: change 4",
		},
	}.String()

	assert.EqualValues(t, expected, actual)
}
