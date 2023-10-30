package webcache

import (
	"net/http"
	"time"
)

type Freshness int

const (
	FreshnessFresh Freshness = iota
	FreshnessStale
	FreshnesTransparent
)

// freshnessFromMaxAge returns the freshness of the response based on the max-age value.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#fresh_and_stale_based_on_age
func freshnessFromMaxAge(maxAge int, responseDated time.Time) Freshness {
	if maxAge < 0 {
		return FreshnessStale
	}
	if maxAge == 0 {
		return FreshnesTransparent
	}
	if responseDated.IsZero() {
		return FreshnessStale
	}
	if time.Now().After(responseDated.Add(time.Duration(maxAge) * time.Second)) {
		return FreshnessStale
	}
	return FreshnessFresh
}

// freshnessFromExpire returns the freshness of the response based on the expire value.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#expires_or_max-age
func freshnessFromExpire(expireTime time.Time, responseDated time.Time) Freshness {
	if expireTime.IsZero() {
		return FreshnesTransparent
	}
	if expireTime.Before(responseDated) {
		return FreshnessStale
	}
	return FreshnessFresh
}

// freshnessFromAge returns the freshness of the response based on the age value.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#fresh_and_stale_based_on_age
func freshnessFromAge(age int, maxAge int, responseDated time.Time) Freshness {
	return freshnessFromMaxAge(maxAge-age, responseDated)
}

// steps to check the freshness of a response:
// 1. check if the response is cacheable
// 2. check if the response is fresh based on max-age
// 3. check if the response is fresh based on expires
// 4. check if the response is fresh based on age
// 5. if none of the above, the response is stale

type FreshnessChecker interface {
	Check(header http.Header, cacheControlHeader CacheControl) (Freshness, error)
}

func NewFreshnerChecker() FreshnessChecker {
	return maxAgeFreshnessChecker{
		expireFreshnessChecker{
			ageFreshnessChecker{
				transparentFreshnessChecker{},
			},
		},
	}
}

type maxAgeFreshnessChecker struct {
	next FreshnessChecker
}

func (c maxAgeFreshnessChecker) Check(header http.Header, cacheControlHeader CacheControl) (Freshness, error) {
	maxAge, err := cacheControlHeader.MaxAge()
	if err != nil {
		return c.next.Check(header, cacheControlHeader)
	}

	date, err := dateFromHeader(header)
	if err != nil {
		return c.next.Check(header, cacheControlHeader)
	}

	return freshnessFromMaxAge(maxAge, date), nil
}

type expireFreshnessChecker struct {
	next FreshnessChecker
}

func (c expireFreshnessChecker) Check(header http.Header, cacheControlHeader CacheControl) (Freshness, error) {
	expires, err := expiresFromHeader(header)
	if err != nil {
		return c.next.Check(header, cacheControlHeader)
	}

	date, err := dateFromHeader(header)
	if err != nil {
		return c.next.Check(header, cacheControlHeader)
	}

	return freshnessFromExpire(expires, date), nil
}

type ageFreshnessChecker struct {
	next FreshnessChecker
}

func (c ageFreshnessChecker) Check(header http.Header, cacheControlHeader CacheControl) (Freshness, error) {

	maxAge, err := cacheControlHeader.MaxAge()
	if err != nil {
		return c.next.Check(header, cacheControlHeader)
	}

	age, err := ageFromHeader(header)
	if err != nil {
		return c.next.Check(header, cacheControlHeader)
	}

	date, err := dateFromHeader(header)
	if err != nil {
		return c.next.Check(header, cacheControlHeader)
	}

	return freshnessFromAge(age, maxAge, date), nil
}

type transparentFreshnessChecker struct{}

func (c transparentFreshnessChecker) Check(header http.Header, cacheControlHeader CacheControl) (Freshness, error) {
	return FreshnesTransparent, nil
}
