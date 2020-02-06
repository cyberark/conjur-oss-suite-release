package changelog

import (
	"fmt"
	"sort"
	"strings"
)

type CombinedChangelog map[string][]string

func (c CombinedChangelog) String() string {
	res := ""

	for section, sectionValues := range c {
		if section == "_" {
			continue
		}

		res = res + fmt.Sprintf("### %s\n", section)
		for _, sectionValue := range sectionValues {
			res = res + fmt.Sprintf("- %s\n", sectionValue)
		}

		res = res + "\n"
	}

	return res
}

func NewCombinedChangelog(changelogs ...*VersionChangelog) CombinedChangelog  {
	res := CombinedChangelog{}

	sort.Slice(changelogs, func(i, j int) bool {
		return changelogs[i].Repo < changelogs[j].Repo
	})

	for _, changelog := range changelogs {
		for section, sectionValues := range changelog.Sections {
			// normalise section keys
			section = strings.ToUpper(section)

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
