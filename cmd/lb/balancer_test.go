package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestScheme(t *testing.T) {
	t.Run("HTTP", func(t *testing.T) {
		*https = false
		assert.Equal(t, "http", scheme())
	})

	t.Run("HTTPS", func(t *testing.T) {
		*https = true
		assert.Equal(t, "https", scheme())
	})
}

func TestHash(t *testing.T) {
	t.Run("Different inputs should produce different hashes", func(t *testing.T) {
		assert.NotEqual(t, hash("test1"), hash("test2"))
		assert.NotEqual(t, hash("abc"), hash("def"))
		assert.NotEqual(t, hash("123"), hash("456"))
		assert.NotEqual(t, hash("test"), hash("test123"))
	})

	t.Run("Same inputs should produce same hashes", func(t *testing.T) {
		assert.Equal(t, hash("test1"), hash("test1"))
		assert.Equal(t, hash("abc"), hash("abc"))
		assert.Equal(t, hash("123"), hash("123"))
		assert.Equal(t, hash(""), hash(""))
	})
}

type MockHttpClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestHealth(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	dst := "server1:8080"
	*https = false

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://%s/health", dst),
		httpmock.NewStringResponder(200, ""))

	client := &MockHttpClient{}

	assert.True(t, health(dst, client))

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://%s/health", dst),
		httpmock.NewStringResponder(500, ""))

	assert.False(t, health(dst, client))
}

func TestHealth_Non200Status(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	dst := "server1:8080"
	*https = false

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://%s/health", dst),
		httpmock.NewStringResponder(503, ""))

	client := &MockHttpClient{}

	assert.False(t, health(dst, client))

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://%s/health", dst),
		httpmock.NewStringResponder(404, ""))

	assert.False(t, health(dst, client))
}

func TestHealth_RequestError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	dst := "server1:8080"
	*https = false

	expectedErr := fmt.Errorf("request error")

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://%s/health", dst),
		func(req *http.Request) (*http.Response, error) {
			return nil, expectedErr
		})

	client := &MockHttpClient{}

	assert.False(t, health(dst, client))
}

func TestHealth_RequestTimeout(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	dst := "server1:8080"
	*https = false

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://%s/health", dst),
		func(req *http.Request) (*http.Response, error) {
			// Sleep longer than the timeout to simulate a timeout error
			time.Sleep(time.Second * 2)
			return nil, context.DeadlineExceeded
		})

	client := &MockHttpClient{}

	assert.False(t, health(dst, client))
}

func TestForward(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	dst := "server1:8080"
	httpmock.RegisterResponder("GET", fmt.Sprintf("http://%s/test", dst),
		httpmock.NewStringResponder(200, "test"))

	
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)
	rw := httptest.NewRecorder()
	client := http.DefaultClient

	err = forward(dst, rw, req, client)
	assert.NoError(t, err)
}