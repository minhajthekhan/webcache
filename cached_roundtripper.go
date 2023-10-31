package webcache

import "net/http"

type httpCacheRoundTripper struct {
	cache            HTTPCache
	next             http.RoundTripper
	freshnessChecker freshnessChecker
}

// NewRoundTripper
func NewRoundTripper(cache Cache, next http.RoundTripper) http.RoundTripper {
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
			response.Header = withCacheHitHeader(response.Header)
			return response, nil
		}
	}
	return nil, nil
}
