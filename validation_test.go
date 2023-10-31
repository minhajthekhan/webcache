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

func TestShouldReturnOriginalResposneIfNoLastModifiedHeader(t *testing.T) {
	roundTripper := &mockRoundTripper{
		statusCode: http.StatusNotModified,
		body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}
	validator := newIfLastModifiedResponseValidator(roundTripper)
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
	validator := newIfLastModifiedResponseValidator(roundTripper)

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
	statusCode int
	body       io.ReadCloser
}

func (m *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusNotModified,
		Body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}, nil
}
