package changelog

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type VersionChangelog struct {
	Repo     string
	Version  string
	Title    string
	Date     string
	Body     string
	Sections map[string][]string
}

// semantic versioning pattern
// [a.b.c], e.g. 2.2.3-pre.1, 2.0.0-x.7.z.92, v1.3.0.
const semverRgx = `\[?v?([\w\d.-]+\.[\w\d.-]+[a-zA-Z0-9])\]?`

// date pattern.
// YYYY-MM-DD (or DD.MM.YYYY, D/M/YY, etc.)
const dateRgx = `.*[ ](\d\d?\d?\d?[-/.]\d\d?[-/.]\d\d?\d?\d?).*`

// link label pattern
// [a.b.c]: http://altavista.com
const linkLabelRgx = `^\[[^[\]]*\] *?:`

// version line pattern
// ## x.y.z - YYYY-MM-DD (or DD.MM.YYYY, D/M/YY, etc.)
const versionLineRgx = `^##? ?[^#]`

// subhead pattern
// ### meow
// #### moo
const subhead = "^###"

// list item pattern
// * list item 1
// * list item 2
const listitem = "^[*-]"

// Parse extracts and returns a slice of changelogs, one for each version.
// Parse assumes a changelog structured roughly like so:
//
// # changelog title
//
// A cool description (optional).
//
// ## unreleased
// * foo
//
// ## x.y.z - YYYY-MM-DD (or DD.MM.YYYY, D/M/YY, etc.)
// * bar
//
// ## [a.b.c]
//
// ### Changes
//
// * Update API
// * Fix bug #1
//
// ## 2.2.3-pre.1 - 2013-02-14
// * Update API
//
// ## 2.0.0-x.7.z.92 - 2013-02-14
// * bark bark
// * woof
// * arf
//
// ## v1.3.0
//
// * make it so
//
// ## [1.2.3](link)
// * init
//
// [a.b.c]: http://altavista.com
func Parse(repo string, changelog string) ([]*VersionChangelog, error) {
	scanner := bufio.NewScanner(strings.NewReader(changelog))

	var versionChangelog *VersionChangelog
	var changeLogs []*VersionChangelog
	var activeSubheader string

	for scanner.Scan() {
		line := scanner.Text()

		// skip line if it's a link label!
		if match, _ := regexp.MatchString(linkLabelRgx, line); match {
			continue
		}

		// new version found!
		if match, _ := regexp.MatchString(versionLineRgx, line); match {
			hasPendingChangelog := versionChangelog != nil
			hasPendingChangelog = hasPendingChangelog && versionChangelog.Title != ""
			hasPendingChangelog = hasPendingChangelog && versionChangelog.Version != ""
			if hasPendingChangelog {
				changeLogs = append(changeLogs, versionChangelog)
			}

			versionChangelog = &VersionChangelog{
				Repo: repo,
				Sections: map[string][]string{
					"_": {},
				},
			}
			activeSubheader = ""

			// extract title
			versionChangelog.Title = strings.TrimSpace(string(line)[2:])

			// extract version
			version := regexp.MustCompile(semverRgx).FindStringSubmatch(line)
			if len(version) == 0 {
				continue
			}
			versionChangelog.Version = string(version[1])

			// extract date
			date := regexp.MustCompile(dateRgx).FindStringSubmatch(line)
			if len(date) == 0 {
				continue
			}
			versionChangelog.Date = string(date[1])

			continue
		}

		// accumulate pending changelog's body
		if versionChangelog != nil && strings.TrimSpace(line) != "" {
			versionChangelog.Body += fmt.Sprintln(strings.TrimSpace(line))

			if match := regexp.MustCompile(subhead).MatchString(line); match {
				key := strings.TrimSpace(strings.Replace(line, "###", "", 1))

				if _, ok := versionChangelog.Sections[key]; !ok {
					versionChangelog.Sections[key] = []string{}
					activeSubheader = key
				}
			}

			if match := regexp.MustCompile(listitem).MatchString(line); match {
				line := regexp.MustCompile(listitem).ReplaceAllString(line, "")
				line = strings.TrimSpace(line)

				versionChangelog.Sections["_"] = append(
					versionChangelog.Sections["_"],
					line,
				)

				if activeSubheader != "" {
					versionChangelog.Sections[activeSubheader] = append(
						versionChangelog.Sections[activeSubheader],
						line,
					)
				}
			}
		}
	}

	hasPendingChangelog := versionChangelog != nil
	hasPendingChangelog = hasPendingChangelog && versionChangelog.Title != ""
	hasPendingChangelog = hasPendingChangelog && versionChangelog.Version != ""
	if hasPendingChangelog {
		changeLogs = append(changeLogs, versionChangelog)
	}

	err := scanner.Err()

	return changeLogs, err
}
