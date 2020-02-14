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
		rw.Write([]byte("Page Content"))
	}))
	defer server.Close()

	content, err := GetWithOptions(server.URL+testPath, server.Client())
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, []byte("Page Content"), content)
}

func TestHttpClientGetRequestUrlProblem(t *testing.T) {
	_, err := Get("zzz")
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err, "Get zzz: unsupported protocol scheme \"\"")
}

func TestHttpClientGetRequestBadStatusCode(t *testing.T) {
	testPath := "/foo/bar"

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(404)
	}))
	defer server.Close()

	_, err := GetWithOptions(server.URL+testPath, server.Client())
	if !assert.Error(t, err) {
		return
	}

	assert.EqualError(t, err, "code 404: "+server.URL+testPath+": ")
}
