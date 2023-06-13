package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestForwardSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	mockClient := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				StatusCode:    http.StatusOK,
				Body:          io.NopCloser(strings.NewReader("mock response")),
				Header:        make(http.Header),
				ContentLength: int64(len("mock response")),
				Request:       req,
			}
			return resp, nil
		},
	}
	err := forward("localhost:8080", rr, req, mockClient)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "mock response", rr.Body.String())
}

func TestForwardFailure(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	errExpected := errors.New("mock error")
	mockClient := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errExpected
		},
	}

	err := forward("localhost:8080", rr, req, mockClient)

	require.Error(t, err)
	assert.Equal(t, errExpected, err)
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestForwardTimeout(t *testing.T) {
	client := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			time.Sleep(2 * time.Second)
			return nil, errors.New("timeout")
		},
	}

	req, err := http.NewRequest("GET", "http://localhost", nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	err = forward("localhost", rr, req, client)
	require.Error(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}
