package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpClientGet(t *testing.T) {
	testPath := "/foo/bar"

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), testPath)
		assert.Equal(t, req.Header.Get("Authorization"), "")
		rw.Write([]byte("Page Content"))
	}))
	defer server.Close()

	client := NewClient()
	content, err := client.Get(server.URL + testPath)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, []byte("Page Content"), content)
}

func TestHttpClientGetTokenSupport(t *testing.T) {
	testPath := "/foo/bar"
	githubToken := "myapikey"

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), testPath)
		assert.Equal(t, req.Header.Get("Authorization"), "token "+githubToken)
		rw.Write([]byte("Page Content"))
	}))
	defer server.Close()

	client := NewClient()
	client.AuthToken = githubToken

	content, err := client.Get(server.URL + testPath)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, []byte("Page Content"), content)
}

func TestHttpClientGetRequestUrlProblem(t *testing.T) {
	client := NewClient()
	_, err := client.Get("zzz")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err, "Get \"zzz\": unsupported protocol scheme \"\"")
}

func TestHttpClientGetRequestBadStatusCode(t *testing.T) {
	testPath := "/foo/bar"

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(404)
	}))
	defer server.Close()

	client := NewClient()
	_, err := client.Get(server.URL + testPath)
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err, "code 404: "+server.URL+testPath+": ")
}
