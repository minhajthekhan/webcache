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

type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, response V)
	Delete(key K)
}

func NewCache() Cache[string, []byte] {
	return &cache{}
}

func (c *cache) Get(key string) ([]byte, bool) {
	v, ok := c.store.Load(key)
	if !ok {
		return nil, false
	}
	return v.([]byte), true
}

func (c *cache) Set(key string, value []byte) {
	c.store.Store(key, value)
}

func (c *cache) Delete(key string) {
	c.store.Delete(key)
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
