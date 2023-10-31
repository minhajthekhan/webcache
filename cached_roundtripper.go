package webcache

import "net/http"

type Transport struct {
	clock            Clock
	cache            HTTPCache
	next             http.RoundTripper
	freshnessChecker freshnessChecker
}

type RoundTripperOption struct {
	Clock Clock
}

type TransportOption func(*Transport)

// NewRoundTripper
func NewTransport(cache Cache, next http.RoundTripper, opts ...TransportOption) *Transport {
	t := &Transport{
		cache: NewHTTPCache(cache),
		next:  next,
		clock: NewClock(),
	}
	for _, o := range opts {
		o(t)
	}
	t.freshnessChecker = newFreshnerChecker(t.clock)
	return t
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	// check if we have this request in the cache
	ctx := r.Context()
	response, ok := t.cache.Get(r)
	// if it does exist in the cache
	if ok {
		cacheControlHeaders := newCacheControl(response.Header)

		// we check if the response is still fresh, if it is, we return it
		freshness, err := t.freshnessChecker.Freshness(ctx, response.Header, cacheControlHeaders)
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
