package webcache

import "net/http"

type httpCacheRoundTripper struct {
	cache            HTTPCache
	next             http.RoundTripper
	freshnessChecker freshnessChecker
}

type RoundTripperOption struct {
	Clock Clock
}

type RoundTripperOptionFunc func(*RoundTripperOption)

// NewRoundTripper
func NewRoundTripper(cache Cache, next http.RoundTripper, options *RoundTripperOption) http.RoundTripper {

	clock := NewClock()
	if options != nil {
		if options.Clock != nil {
			clock = options.Clock
		}
	}

	return &httpCacheRoundTripper{
		cache:            NewHTTPCache(cache),
		next:             next,
		freshnessChecker: newFreshnerChecker(clock),
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
