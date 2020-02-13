package github

type ReleaseInfo struct {
	Description string `json:"body"`
	Draft       bool   `json:"draft"`
	Name        string `json:"name"`
	TagName     string `json:"tag_name"`
}
