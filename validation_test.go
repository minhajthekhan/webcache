package webcache

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// if the response from the cash is stale, the client should send a conditional request to the origin server
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#validation

func TestResponseValidationShouldNotCacheIfNoHeadersPresent(t *testing.T) {
	roundTripper := &mockRoundTripper{
		statusCode: http.StatusNotModified,
		body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}
	validator := newResponseValidator(roundTripper)
	responseHeaders := make(http.Header)
	cachedResponse := http.Response{
		Header: responseHeaders,
		Body:   io.NopCloser(nil),
	}

	r, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(t, err)

	response, err := validator.Validate(&cachedResponse, r)
	assert.NoError(t, err)
	assert.Equal(t, "", response.Header.Get("X-Cache"))
}

func TestResponseValidationShouldCacheIfNoEtagButLastModified(t *testing.T) {
	lastModified := time.Now().Format(http.TimeFormat)

	roundTripper := &mockRoundTripper{
		statusCode:           http.StatusNotModified,
		body:                 io.NopCloser(bytes.NewReader([]byte(""))),
		assertLastModified:   true,
		testingT:             t,
		ifModifiedSinceValue: lastModified,
	}

	validator := newResponseValidator(roundTripper)
	responseHeaders := make(http.Header)
	responseHeaders.Add("Last-Modified", lastModified)
	cachedResponse := http.Response{
		Header: responseHeaders,
		Body:   io.NopCloser(nil),
	}

	r, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(t, err)

	response, err := validator.Validate(&cachedResponse, r)
	assert.NoError(t, err)
	assert.Equal(t, "HIT", response.Header.Get("X-Cache"))
}

func TestResponseValidationShouldCacheIfNoLastModifiedButEtag(t *testing.T) {
	roundTripper := &mockRoundTripper{
		statusCode:        http.StatusNotModified,
		body:              io.NopCloser(bytes.NewReader([]byte(""))),
		assertIfNoneMatch: true,
		testingT:          t,
		ifNoneMatchValue:  "123",
	}
	validator := newResponseValidator(roundTripper)
	responseHeaders := make(http.Header)
	responseHeaders.Add("ETag", "123")
	cachedResponse := http.Response{
		Header: responseHeaders,
		Body:   io.NopCloser(nil),
	}

	r, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(t, err)

	response, err := validator.Validate(&cachedResponse, r)
	assert.NoError(t, err)
	assert.Equal(t, "HIT", response.Header.Get("X-Cache"))
}

func TestShouldReturnOriginalResposneIfNoETagHeader(t *testing.T) {
	roundTripper := &mockRoundTripper{
		statusCode: http.StatusNotModified,
		body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}
	validator := newEtagResponseValidator(roundTripper, newRevalidator(roundTripper))
	responseHeaders := make(http.Header)
	cachedResponse := http.Response{
		Header: responseHeaders,
		Body:   io.NopCloser(nil),
	}

	r, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(t, err)

	response, err := validator.Validate(&cachedResponse, r)
	assert.NoError(t, err)
	assert.Equal(t, "", response.Header.Get("X-Cache"))
}

func TestShouldReturnCachedResponseIfEtagRoundTripperReturns304OnValidation(t *testing.T) {
	// the actual request should return a status 304 - Not Modified
	roundTripper := &mockRoundTripper{
		statusCode: http.StatusNotModified,
		body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}
	validator := newEtagResponseValidator(roundTripper, newRevalidator(roundTripper))

	// the response in the cache can have a last modified header
	responseHeaders := make(http.Header)
	responseHeaders.Add("ETag", "123456789")
	cachedResponse := http.Response{
		Header: responseHeaders,
		Body:   io.NopCloser(nil),
	}

	r, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(t, err)

	response, err := validator.Validate(&cachedResponse, r)
	assert.NoError(t, err)
	assert.Equal(t, "HIT", response.Header.Get("X-Cache"))
}

func TestShouldReturnOriginalResposneIfNoLastModifiedHeader(t *testing.T) {
	roundTripper := &mockRoundTripper{
		statusCode: http.StatusNotModified,
		body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}
	validator := newIfLastModifiedResponseValidator(roundTripper, newRevalidator(roundTripper))
	responseHeaders := make(http.Header)
	cachedResponse := http.Response{
		Header: responseHeaders,
		Body:   io.NopCloser(nil),
	}

	r, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(t, err)

	response, err := validator.Validate(&cachedResponse, r)
	assert.NoError(t, err)
	assert.Equal(t, "", response.Header.Get("X-Cache"))
}
func TestShouldReturnCachedResponseIfRoundTripperReturns304OnValidation(t *testing.T) {
	// the actual request should return a status 304 - Not Modified
	roundTripper := &mockRoundTripper{
		statusCode: http.StatusNotModified,
		body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}
	validator := newIfLastModifiedResponseValidator(roundTripper, newRevalidator(roundTripper))

	// the response in the cache can have a last modified header
	responseHeaders := make(http.Header)
	responseHeaders.Add("Last-Modified", time.Now().Format(http.TimeFormat))
	cachedResponse := http.Response{
		Header: responseHeaders,
		Body:   io.NopCloser(nil),
	}

	r, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(t, err)

	response, err := validator.Validate(&cachedResponse, r)
	assert.NoError(t, err)
	assert.Equal(t, "HIT", response.Header.Get("X-Cache"))
}

type mockRoundTripper struct {
	testingT   *testing.T
	statusCode int
	body       io.ReadCloser

	assertLastModified   bool
	ifModifiedSinceValue string

	assertIfNoneMatch bool
	ifNoneMatchValue  string

	response *http.Response
}

func (m *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {

	if m.assertLastModified {
		assert.Equal(m.testingT, m.ifModifiedSinceValue, r.Header.Get("If-Modified-Since"))
	}
	if m.assertIfNoneMatch {
		assert.Equal(m.testingT, m.ifNoneMatchValue, r.Header.Get("If-None-Match"))
	}
	if m.response != nil {
		return m.response, nil
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       m.body,
	}, nil
}
