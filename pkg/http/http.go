package http

import (
	"fmt"
	"io/ioutil"
	stdlibHttp "net/http"

	"github.com/cyberark/conjur-oss-suite-release/pkg/log"
)

// Client is a wrapper around stdlibHttp client but with added storage for
// an auth token
type Client struct {
	*stdlibHttp.Client
	AuthToken string
}

// NewClient creates a Client with an initialized parent stdlibHttp.Client
// object
func NewClient() *Client {
	return &Client{
		&stdlibHttp.Client{},
		"",
	}
}

// Get retrieves the content of a URL
func (client *Client) Get(url string) ([]byte, error) {
	request, err := stdlibHttp.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add API auth token if one is provided
	if client.AuthToken != "" {
		request.Header.Add("Authorization", "token "+client.AuthToken)
	}

	log.OutLogger.Printf("  Fetching %s...", url)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("code %d: %s: %s", response.StatusCode, url, contents)
	}

	return contents, nil
}
