package changelog

import (
	"fmt"
	"sort"
	"strings"
)

// UnifiedChangelog is a map of repos to their changelog sections
type UnifiedChangelog map[string][]string

func (c UnifiedChangelog) String() string {
	res := ""

	var sections []string
	for section := range c {
		sections = append(sections, section)
	}
	sort.Strings(sections)

	for _, section := range sections {
		if section == "_" {
			continue
		}

		sectionValues := c[section]

		res = res + fmt.Sprintf("### %s\n", section)
		for _, sectionValue := range sectionValues {
			res = res + fmt.Sprintf("- %s\n", sectionValue)
		}

		res = res + "\n"
	}

	return res
}

// NewUnifiedChangelog creates a unified changelog from varios per-version and
// per-repo changelogs
func NewUnifiedChangelog(changelogs ...*VersionChangelog) UnifiedChangelog {
	res := UnifiedChangelog{}

	sort.Slice(changelogs, func(i, j int) bool {
		return changelogs[i].Repo < changelogs[j].Repo
	})

	for _, changelog := range changelogs {
		for section, sectionValues := range changelog.Sections {
			// normalise section keys
			section = strings.Title(strings.ToLower(section))

			if section == "_" {
				continue
			}

			if _, ok := res[section]; !ok {
				res[section] = []string{}
			}

			for _, value := range sectionValues {
				res[section] = append(
					res[section],
					fmt.Sprintf(
						"`%s@%s`: %s",
						changelog.Repo,
						changelog.Version,
						value,
					),
				)
			}
		}
	}

	return res
}
