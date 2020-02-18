package github

// ReleaseInfo is a representation of a v3 GitHub API JSON
// structure denitong a release. We only are interested in
// a small subsection of the field so this list is trimmed
// from the full one that the API returns.
type ReleaseInfo struct {
	Description string `json:"body"`
	Draft       bool   `json:"draft"`
	Name        string `json:"name"`
	TagName     string `json:"tag_name"`
}
