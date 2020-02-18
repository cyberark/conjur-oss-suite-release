package changelog

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	markdownparser "github.com/gomarkdown/markdown/parser"
)

// VersionChangelog contains a full changelog for a version of a repo
type VersionChangelog struct {
	Repo     string
	Version  string
	Date     string
	Sections map[string][]string
}

// semantic versioning pattern
// [a.b.c], e.g. 2.2.3-pre.1, 2.0.0-x.7.z.92, v1.3.0.
const semverRgx = `\[?v?([\w\d.-]+\.[\w\d.-]+[a-zA-Z0-9])\]?`

// date pattern.
// YYYY-MM-DD (or DD.MM.YYYY, D/M/YY, etc.)
const dateRgx = `.*[ ](\d\d?\d?\d?[-/.]\d\d?[-/.]\d\d?\d?\d?).*`

// Parse extracts and returns a slice of changelogs, one for each version.
// Parse assumes a changelog in the [keep a changelog](https://keepachangelog.com/) format:
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
	var changelogs []*VersionChangelog

	parser := markdownparser.New()
	rootNode := parser.Parse([]byte(changelog))

	// state-machine state
	var insideVersion = false
	var insideSection = false
	var insideListItem = false

	// Buffers used to extract values spanning multiple AST nodes
	var versionBuffer = ""
	var sectionBuffer = ""
	var listItemBuffer = ""

	// Extract changelog versions
	ast.WalkFunc(rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		switch n := node.(type) {
		// Handle list-item found anywhere along version > section > list-item
		case *ast.ListItem:
			insideListItem = entering

			if entering {
				// On entering list-item node under version section, reset list-item buffer
				listItemBuffer = ""
			} else {
				// On exiting list-item node under version section, populate changelog with
				// list-item by reading list-item buffer
				if _, ok := versionChangelog.Sections[sectionBuffer]; !ok {
					versionChangelog.Sections[sectionBuffer] = []string{}
				}

				versionChangelog.Sections[sectionBuffer] = append(
					versionChangelog.Sections[sectionBuffer],
					string(listItemBuffer),
				)
			}
		// Handle text
		case *ast.Text:
			switch {
			case insideVersion:
				versionBuffer += string(n.Literal)
			case insideSection:
				sectionBuffer += string(n.Literal)
			case insideListItem:
				listItemBuffer += string(n.Literal)
			}
		// Handle link found anywhere along version > section > list-item
		case *ast.Link:
			txt := ""
			if entering {
				txt = "["
			} else {
				txt = fmt.Sprintf("](%s)", string(n.Destination))
			}

			// Populate relevant buffer with relevant component of link's textual
			// representation
			switch {
			case insideSection:
				sectionBuffer += txt
			case insideListItem:
				listItemBuffer += txt
			case insideVersion:
				// Do nothing.
				// This avoids having destination as part of the versionBuffer. This is
				// because the link regex doesn't expect the link destination to be
				// present.
			}
		// Handle heading
		case *ast.Heading:
			switch n.Level {
			// Handle version
			case len("##"):
				insideVersion = entering

				if entering {
					// On entering version header node, initialise changelog
					versionChangelog = &VersionChangelog{
						Repo:     repo,
						Sections: map[string][]string{},
					}
					versionBuffer = ""

					changelogs = append(changelogs, versionChangelog)

				} else {
					// On exiting version header node, populate changelog

					// Extract version
					version := regexp.MustCompile(semverRgx).FindStringSubmatch(versionBuffer)
					if len(version) == 0 {
						break
					}
					versionChangelog.Version = string(version[1])

					// Extract date
					date := regexp.MustCompile(dateRgx).FindStringSubmatch(versionBuffer)
					if len(date) == 0 {
						break
					}
					versionChangelog.Date = string(date[1])
				}

			// Handle section under version
			case len("###"):
				insideSection = entering

				// On entering section node under version header, reset section buffer
				if entering {
					sectionBuffer = ""
				}
			}
		}

		return ast.GoToNext
	})

	// Select valid changelogs only, using in-place filter
	// A valid changelog is one that has a non-empty version.
	n := 0
	for _, changelog := range changelogs {
		if changelog != nil && changelog.Version != "" {
			changelogs[n] = changelog
			n++
		}
	}
	changelogs = changelogs[:n]

	err := scanner.Err()

	return changelogs, err
}
