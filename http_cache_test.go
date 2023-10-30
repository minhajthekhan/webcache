package webcache

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPCache(t *testing.T) {
	cache := NewHTTPCache(NewCache())
	r, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.NoError(t, err)
	cache.Get(r)
}
