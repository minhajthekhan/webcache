package webcache

import "net/http"

type HTTPCache interface {
	Get(r *http.Request) (*http.Response, bool)
	Set(r *http.Request, response *http.Response)
}

type httpCache struct {
	cache Cache
}

func NewHTTPCache(cache Cache) HTTPCache {
	return &httpCache{cache: cache}
}

func (c *httpCache) Get(r *http.Request) (*http.Response, bool) {
	cacheKey := buildCacheKey(r)
	return c.cache.Get(cacheKey)
}

func (c *httpCache) Set(r *http.Request, response *http.Response) {
	cacheKey := buildCacheKey(r)
	c.cache.Set(cacheKey, response)
}

func isCached(r *http.Response) bool {
	return r.Header.Get("X-Cache") == "HIT"
}
