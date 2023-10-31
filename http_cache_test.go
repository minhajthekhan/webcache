package webcache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheKey(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.NoError(t, err)
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Accept-Language", "en-US")
	r.Header.Add("Cache-Control", "max-age=120")
	r.Header.Add("Date", time.Now().Format(time.RFC850))
	r.Header.Add("Vary", "Accept, Accept-Language")
	key := buildCacheKey(r)
	assert.NotEmpty(t, key)
	assert.Equal(t, "cache_key=GET_http://example.com_application/json_en-US", key.String())

	r, err = http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.NoError(t, err)
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Accept-Language", "en-US")
	r.Header.Add("Cache-Control", "max-age=120")
	r.Header.Add("Date", time.Now().Format(time.RFC850))
	r.Header.Add("Vary", "Accept")
	key = buildCacheKey(r)
	assert.NotEmpty(t, key)
	assert.Equal(t, "cache_key=GET_http://example.com_application/json", key.String())

}
