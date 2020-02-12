package github

type ReleaseInfo struct {
	Description string `json:"body"`
	TagName     string `json:"tag_name"`
}
