package webcache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	c := NewCache()
	assert.NotNil(t, c)
	r, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.NoError(t, err)
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Accept-Language", "en-US")
	r.Header.Add("Cache-Control", "max-age=120")
	r.Header.Add("Date", time.Now().Format(time.RFC850))
	r.Header.Add("Vary", "Accept, Accept-Language")

	key := buildCacheKey(r)
	resp, ok := c.Get(key.String())
	assert.False(t, ok)
	assert.Nil(t, resp)

	c.Set(key.String(), []byte("hello world"))
	resp, ok = c.Get(key.String())
	assert.True(t, ok)
	assert.NotNil(t, resp)

	c.Delete(key.String())
	_, ok = c.Get(key.String())
	assert.False(t, ok)
}
