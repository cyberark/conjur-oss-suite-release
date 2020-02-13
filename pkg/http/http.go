package http

import (
	"fmt"
	"io/ioutil"
	"log"
	stdlibHttp "net/http"
)

func Get(url string) ([]byte, error) {
	client := &stdlibHttp.Client{}

	request, err := stdlibHttp.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	log.Printf("  Fetching %s...", url)
	response, err := client.Do(request)

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
