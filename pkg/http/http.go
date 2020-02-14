package http

import (
	"fmt"
	"io/ioutil"
	"log"
	stdlibHttp "net/http"
)

func Get(url string) ([]byte, error) {
	client := &stdlibHttp.Client{}
	return GetWithOptions(url, client)
}

func GetWithOptions(url string, client *stdlibHttp.Client) ([]byte, error) {
	request, err := stdlibHttp.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	log.Printf("  Fetching %s...", url)
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
