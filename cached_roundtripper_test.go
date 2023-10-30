package webcache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type httpCacheRoundTripper struct {
	cache            HTTPCache
	next             http.RoundTripper
	freshnessChecker freshnessChecker
}

func TestRoundTripperIfRequestExistsInCache(t *testing.T) {

	resp := http.Response{Header: make(http.Header), StatusCode: http.StatusOK}
	resp.Header.Set("Cache-Control", "max-age=120")
	resp.Header.Set("Date", time.Now().Format(time.RFC850))

	cache := NewCache()
	r, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	cache.Set(buildCacheKey(r), &resp)

	roundTripper := NewHTTPCacheRoundTripper(cache, http.DefaultTransport)
	assert.NoError(t, err)

	response, err := roundTripper.RoundTrip(r)
	assert.NoError(t, err)
	assert.Equal(t, response.StatusCode, http.StatusOK)
	assert.Equal(t, "HIT", response.Header.Get("X-Cache"))
}

func NewHTTPCacheRoundTripper(cache Cache, next http.RoundTripper) http.RoundTripper {
	return &httpCacheRoundTripper{
		cache:            NewHTTPCache(cache),
		next:             next,
		freshnessChecker: newFreshnerChecker(),
	}
}

func (c *httpCacheRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {

	// check if we have this request in the cache
	response, ok := c.cache.Get(r)
	// if it does exist in the cache
	if ok {
		cacheControlHeaders := newCacheControl(response.Header)

		// we check if the response is still fresh, if it is, we return it
		freshness, err := c.freshnessChecker.Freshness(response.Header, cacheControlHeaders)
		if err != nil {
			return nil, err
		}
		if freshness == FreshnessFresh {
			response.Header.Set("X-Cache", "HIT")
			return response, nil
		}
	}
	return nil, nil
}
