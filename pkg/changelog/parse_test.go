package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const validChangelog = `
# Changelog
description
## [Unreleased]

## [1.5.0] 2020-01-29

### Added
- add 1

### Changed
- change 1
- change 2
`

func TestParse(t *testing.T) {
	changelogs, err := Parse("test-repo", validChangelog)

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
- change 2
`,
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
