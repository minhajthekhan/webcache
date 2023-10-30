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

func TestCacheControl(t *testing.T) {
	header := make(http.Header)
	header.Add("Cache-Control", "max-age=100")
	cc := newCacheControl(header)
	age, _ := cc.MaxAge()
	assert.Equal(t, 100, age)
}
