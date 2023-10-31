package webcache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheControlMaxAge(t *testing.T) {
	cc := CacheControl{cacheControlKeyMaxAge: "100"}
	age, _ := cc.MaxAge()
	assert.Equal(t, 100, age)

	cc = CacheControl{}
	age, _ = cc.MaxAge()
	assert.Equal(t, 0, age)

	cc = CacheControl{cacheControlKeyMaxAge: "abc"}
	age, _ = cc.MaxAge()
	assert.Equal(t, 0, age)
}

func TestDateFromHeader(t *testing.T) {

	_, err := dateFromHeader(nil)
	assert.Error(t, err)

	_, err = dateFromHeader(http.Header{})
	assert.Error(t, err)

	_, err = dateFromHeader(http.Header{"Date": []string{}})
	assert.Error(t, err)

	_, err = dateFromHeader(http.Header{"Date": []string{"abc"}})
	assert.Error(t, err)

	expected := time.Now().Format(time.RFC850)
	date, err := dateFromHeader(http.Header{"Date": []string{expected}})
	assert.NoError(t, err)
	assert.Equal(t, expected, date.Format(time.RFC850))
}

func TestExpiresFromHeader(t *testing.T) {
	_, err := expiresFromHeader(nil)
	assert.Error(t, err)

	_, err = expiresFromHeader(http.Header{})
	assert.Error(t, err)

	_, err = expiresFromHeader(http.Header{"Expires": []string{}})
	assert.Error(t, err)

	_, err = expiresFromHeader(http.Header{"Expires": []string{"abc"}})
	assert.Error(t, err)

	expected := time.Now().Format(time.RFC850)
	date, err := expiresFromHeader(http.Header{"Expires": []string{expected}})
	assert.NoError(t, err)
	assert.Equal(t, expected, date.Format(time.RFC850))
}

func TestAgeFromHeader(t *testing.T) {
	_, err := ageFromHeader(nil)
	assert.Error(t, err)

	_, err = ageFromHeader(http.Header{})
	assert.Error(t, err)

	_, err = ageFromHeader(http.Header{"Age": []string{}})
	assert.Error(t, err)

	_, err = ageFromHeader(http.Header{"Age": []string{"abc"}})
	assert.Error(t, err)

	age, err := ageFromHeader(http.Header{"Age": []string{"100"}})
	assert.NoError(t, err)
	assert.Equal(t, 100, age)
}

func TestLastModifiedHeader(t *testing.T) {

	_, err := lastModifiedFromHeader(http.Header{"Age": []string{"abc"}})
	assert.Error(t, err)

	lastMod, err := lastModifiedFromHeader(http.Header{"Last-Modified": []string{time.Now().Format(http.TimeFormat)}})
	assert.NoError(t, err)
	assert.Equal(t, time.Now().Format(http.TimeFormat), lastMod.Format(http.TimeFormat))
}

func TestEtagFromHeader(t *testing.T) {
	_, err := etagFromHeader(http.Header{})
	assert.Error(t, err)

	_, err = etagFromHeader(http.Header{"Etag": []string{}})
	assert.Error(t, err)

	etag, err := etagFromHeader(http.Header{"Etag": []string{"abc"}})
	assert.NoError(t, err)
	assert.Equal(t, "abc", etag)
}

func TestCacheControl(t *testing.T) {
	header := make(http.Header)
	header.Add("Cache-Control", "max-age=100, public")
	cc := newCacheControl(header)

	age, _ := cc.MaxAge()
	assert.Equal(t, 100, age)
	assert.True(t, cc.Public())
	assert.False(t, cc.Private())

	header = make(http.Header)
	header.Add("Cache-Control", "max-age=100, private")
	cc = newCacheControl(header)
	age, _ = cc.MaxAge()
	assert.Equal(t, 100, age)
	assert.False(t, cc.Public())
	assert.True(t, cc.Private())

	header = make(http.Header)
	header.Add("Cache-Control", "max-age=100, no-cache")
	cc = newCacheControl(header)
	assert.True(t, cc.NoCache())

	header = make(http.Header)
	header.Add("Cache-Control", "max-age=100, must-revalidate")
	cc = newCacheControl(header)
	assert.True(t, cc.MustRevalidate())

	header = make(http.Header)
	header.Add("Cache-Control", "max-age=100, must-revalidate, no-store")
	cc = newCacheControl(header)
	assert.True(t, cc.NoStore())

}
