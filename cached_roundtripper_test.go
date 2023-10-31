package webcache

import (
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

// TODO:: add tests for transparent freshness
// TODO:: add test for stale freshness when the response was not modified and can be re-used
// TODO:: add test for stale freshness when the response was modified and can not be re-used
