package webcache

import (
	"net/http"
	"strconv"
	"time"
)

type cacheControlKey string

var (
	cacheControlKeyMaxAge = cacheControlKey("max-age")
)

type CacheControl map[cacheControlKey]string

func (c CacheControl) MaxAge() int {
	v, ok := c[cacheControlKeyMaxAge]
	if !ok {
		return 0
	}
	maxAge, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return maxAge
}

func ageFromHeader(h http.Header) int {
	age, err := strconv.Atoi(h.Get("Age"))
	if err != nil {
		return 0
	}
	return age
}

func expiresFromHeader(h http.Header) time.Time {
	return timeFromHeader(h, "Expires")
}

func dateFromHeader(h http.Header) time.Time {
	return timeFromHeader(h, "Date")
}
func timeFromHeader(h http.Header, key string) time.Time {
	v, err := http.ParseTime(h.Get(key))
	if err != nil {
		return time.Time{}
	}
	return v
}
