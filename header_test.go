package webcache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheControlMaxAge(t *testing.T) {
	cc := CacheControl{cacheControlKeyMaxAge: "100"}
	assert.Equal(t, 100, cc.MaxAge())

	cc = CacheControl{}
	assert.Equal(t, 0, cc.MaxAge())

	cc = CacheControl{cacheControlKeyMaxAge: "abc"}
	assert.Equal(t, 0, cc.MaxAge())
}

func TestDateFromHeader(t *testing.T) {
	assert.Equal(t, time.Time{}, dateFromHeader(nil))
	assert.Equal(t, time.Time{}, dateFromHeader(http.Header{}))
	assert.Equal(t, time.Time{}, dateFromHeader(http.Header{"Date": []string{}}))
	assert.Equal(t, time.Time{}, dateFromHeader(http.Header{"Date": []string{"abc"}}))

	expected, _ := time.Parse(time.RFC1123, "Tue, 22 Feb 2022 22:22:22 UTC")
	assert.Equal(t, expected, dateFromHeader(http.Header{"Date": []string{"Tue, 22 Feb 2022 22:22:22 GMT"}}))
}

func TestExpiresFromHeader(t *testing.T) {
	assert.Equal(t, time.Time{}, expiresFromHeader(nil))
	assert.Equal(t, time.Time{}, expiresFromHeader(http.Header{}))
	assert.Equal(t, time.Time{}, expiresFromHeader(http.Header{"Expires": []string{}}))
	assert.Equal(t, time.Time{}, expiresFromHeader(http.Header{"Expires": []string{"abc"}}))

	expected, _ := time.Parse(time.RFC1123, "Tue, 22 Feb 2022 22:22:22 UTC")
	assert.Equal(t, expected, expiresFromHeader(http.Header{"Expires": []string{"Tue, 22 Feb 2022 22:22:22 GMT"}}))
}

func TestAgeFromHeader(t *testing.T) {
	assert.Equal(t, 0, ageFromHeader(nil))
	assert.Equal(t, 0, ageFromHeader(http.Header{}))
	assert.Equal(t, 0, ageFromHeader(http.Header{"Age": []string{}}))
	assert.Equal(t, 0, ageFromHeader(http.Header{"Age": []string{"abc"}}))
	assert.Equal(t, 100, ageFromHeader(http.Header{"Age": []string{"100"}}))
}
