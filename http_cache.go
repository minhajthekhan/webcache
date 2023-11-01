package webcache

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httputil"
)

type HTTPCache interface {
	Get(r *http.Request) (*http.Response, bool)
	Set(r *http.Request, response *http.Response)
	Delete(r *http.Request)
}

type httpCache struct {
	cache Cache[string, []byte]
}

func NewHTTPCache(cache Cache[string, []byte]) HTTPCache {
	return &httpCache{cache: cache}
}

func (c *httpCache) Get(r *http.Request) (*http.Response, bool) {
	cacheKey := buildCacheKey(r)
	if cachedVal, ok := c.cache.Get(cacheKey.String()); ok {
		b := bytes.NewBuffer(cachedVal)
		v, err := http.ReadResponse(bufio.NewReader(b), nil)
		if err != nil {
			return nil, false
		}
		return v, true
	}
	return nil, false
}

func (c *httpCache) Set(r *http.Request, response *http.Response) {
	b, err := httputil.DumpResponse(response, true)
	if err != nil {
		return
	}

	cacheKey := buildCacheKey(r)
	c.cache.Set(cacheKey.String(), b)
}

func (c *httpCache) Delete(r *http.Request) {
	cacheKey := buildCacheKey(r)
	c.cache.Delete(cacheKey.String())
}

func isCached(r *http.Response) bool {
	return r.Header.Get("X-Cache") == "HIT"
}
