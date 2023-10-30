package webcache

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type cache struct {
	store sync.Map
}

type Cache interface {
	Get(key cacheKey) (*http.Response, bool)
	Set(key cacheKey, response *http.Response)
}

func NewCache() Cache {
	return &cache{}
}

func (c *cache) Get(key cacheKey) (*http.Response, bool) {
	v, ok := c.store.Load(key)
	if !ok {
		return nil, false
	}
	return v.(*http.Response), true
}

func (c *cache) Set(key cacheKey, response *http.Response) {
	c.store.Store(key, response)
}

type cacheKey string

func (k cacheKey) String() string {
	return string(k)
}

// cacheKey returns a cache key for the request.
// The cache key is a string that uniquely identifies the request.
// The cache key is used to store and retrieve the response from the cache.
// The cache key is generated from the request method, the request URL and the Vary header.
func buildCacheKey(r *http.Request) cacheKey {
	components := make([]string, 0)
	components = append(components, r.Method)
	components = append(components, r.URL.String())
	components = append(components, componentsFromVaryHeader(r.Header)...)

	return cacheKey(fmt.Sprintf("cache_key=%s", strings.Join(components, "_")))
}

func componentsFromVaryHeader(h http.Header) []string {
	components := make([]string, 0)
	vary := h.Get("Vary")
	varyHeaderKeys := strings.Split(vary, ",")
	for i := range varyHeaderKeys {
		v := h.Get(strings.TrimSpace(varyHeaderKeys[i]))
		if v != "" {
			components = append(components, v)
		}
	}
	return components
}
