package webcache

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransportIfRequestExistsInCache(t *testing.T) {
	resp := http.Response{Header: make(http.Header), StatusCode: http.StatusOK}
	resp.Header.Set("Cache-Control", "max-age=120")
	resp.Header.Set("Date", time.Now().Format(time.RFC850))

	cache := NewCache()
	r, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	cache.Set(buildCacheKey(r), &resp)

	roundTripper := NewTransport(cache, http.DefaultTransport, WithClock(NewClock()))
	assert.NoError(t, err)

	response, err := roundTripper.RoundTrip(r)
	assert.NoError(t, err)
	assert.Equal(t, response.StatusCode, http.StatusOK)
	assert.Equal(t, "HIT", response.Header.Get("X-Cache"))
}

func TestTransportIfRequestIsStaleWithLastModified(t *testing.T) {
	cache := NewCache()
	headers := make(http.Header)
	headers.Add("Cache-Control", "max-age=0")
	headers.Add("Date", time.Now().Add(-1*time.Minute).Format(http.TimeFormat))
	lastModified := time.Now().Add(-2 * time.Minute).Format(http.TimeFormat)
	headers.Add("Last-Modified", lastModified)

	mockRt := &mockRoundTripper{
		testingT:             t,
		statusCode:           http.StatusNotModified,
		body:                 io.NopCloser(bytes.NewReader([]byte(""))),
		assertLastModified:   true,
		ifModifiedSinceValue: lastModified,
	}

	cachedResponse := http.Response{Header: headers, StatusCode: http.StatusOK}
	r, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.NoError(t, err)

	cache.Set(buildCacheKey(r), &cachedResponse)
	rt := NewTransport(cache, mockRt, WithClock(NewClock()))

	response, err := rt.RoundTrip(r)
	assert.NoError(t, err)
	assert.Equal(t, "HIT", response.Header.Get("X-Cache"))
}

func TestTransportIfRequestIsStaleWithEtag(t *testing.T) {
	cache := NewCache()
	headers := make(http.Header)
	headers.Add("Cache-Control", "max-age=0")
	headers.Add("Date", time.Now().Add(-1*time.Minute).Format(http.TimeFormat))
	etag := "123"
	headers.Add("Etag", etag)

	mockRt := &mockRoundTripper{
		testingT:          t,
		statusCode:        http.StatusNotModified,
		body:              io.NopCloser(bytes.NewReader([]byte(""))),
		assertIfNoneMatch: true,
		ifNoneMatchValue:  "123",
	}

	cachedResponse := http.Response{Header: headers, StatusCode: http.StatusOK}
	r, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.NoError(t, err)

	cache.Set(buildCacheKey(r), &cachedResponse)
	rt := NewTransport(cache, mockRt, WithClock(NewClock()))

	response, err := rt.RoundTrip(r)
	assert.NoError(t, err)
	assert.Equal(t, "HIT", response.Header.Get("X-Cache"))
}

func TestTransportIfRequestIsStaleWithEtagChanged(t *testing.T) {
	cache := NewCache()
	headers := make(http.Header)
	headers.Add("Cache-Control", "max-age=0")
	headers.Add("Date", time.Now().Add(-1*time.Minute).Format(http.TimeFormat))
	etag := "123"
	headers.Add("Etag", etag)

	mockRt := &mockRoundTripper{
		testingT:          t,
		statusCode:        http.StatusOK,
		body:              io.NopCloser(bytes.NewReader([]byte(""))),
		assertIfNoneMatch: false,
		ifNoneMatchValue:  "345",
	}

	cachedResponse := http.Response{Header: headers, StatusCode: http.StatusOK}
	r, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.NoError(t, err)

	cache.Set(buildCacheKey(r), &cachedResponse)
	rt := NewTransport(cache, mockRt, WithClock(NewClock()))
	response, err := rt.RoundTrip(r)
	assert.NoError(t, err)
	assert.Equal(t, "", response.Header.Get("X-Cache"))
}

func TestTransportShouldNotCacheIfNoStoreCacheControlHeader(t *testing.T) {
	cache := NewCache()
	headers := make(http.Header)
	headers.Add("Cache-Control", "max-age=0, no-store")
	headers.Add("Date", time.Now().Add(-1*time.Minute).Format(http.TimeFormat))
	etag := "123"
	headers.Add("Etag", etag)

	mockRt := &mockRoundTripper{
		testingT:          t,
		statusCode:        http.StatusNotModified,
		body:              io.NopCloser(bytes.NewReader([]byte(""))),
		assertIfNoneMatch: true,
		ifNoneMatchValue:  "123",
	}

	cachedResponse := http.Response{Header: headers, StatusCode: http.StatusOK}
	r, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.NoError(t, err)

	cache.Set(buildCacheKey(r), &cachedResponse)
	rt := NewTransport(cache, mockRt, WithClock(NewClock()))

	response, err := rt.RoundTrip(r)
	assert.NoError(t, err)
	assert.Equal(t, "HIT", response.Header.Get("X-Cache"))

	// there should be no cache entry because of the no-store directive
	_, ok := cache.Get(buildCacheKey(r))
	assert.False(t, ok)
}
