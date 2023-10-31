package webcache

import (
	"context"
	"net/http"
	"time"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func NewClock() Clock {
	return realClock{}
}

func (c realClock) Now() time.Time {
	return time.Now()
}

type Freshness int

const (
	FreshnessFresh Freshness = iota
	FreshnessStale
	FreshnesTransparent
)

// freshnessFromMaxAge returns the freshness of the response based on the max-age value.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#fresh_and_stale_based_on_age
func freshnessFromMaxAge(maxAge int, responseDated time.Time, clock Clock) Freshness {
	if maxAge < 0 {
		return FreshnessStale
	}
	if maxAge == 0 {
		return FreshnesTransparent
	}
	if responseDated.IsZero() {
		return FreshnessStale
	}
	if clock.Now().After(responseDated.Add(time.Duration(maxAge) * time.Second)) {
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
func freshnessFromAge(age int, maxAge int) Freshness {
	// 150 - 100
	if maxAge-age > 0 {
		return FreshnessFresh
	}
	return FreshnessStale
}

type freshnessChecker interface {
	Freshness(ctx context.Context, header http.Header, cacheControlHeader CacheControl) (Freshness, error)
}

// Steps to check the freshness of a response:
// 1. check if the response is fresh based on age and max-age
// 2. check if the response is fresh based on max-age
// 3. check if the response is fresh based on expires
// 4. if none of the above, the response is stale
func newFreshnerChecker(clock Clock) freshnessChecker {
	return ageFreshnessChecker{
		maxAgeFreshnessChecker{
			next: expireFreshnessChecker{
				transparentFreshness{},
			},
			clock: clock,
		},
	}
}

type maxAgeFreshnessChecker struct {
	next  freshnessChecker
	clock Clock
}

func (c maxAgeFreshnessChecker) Freshness(ctx context.Context, header http.Header, cacheControlHeader CacheControl) (Freshness, error) {
	maxAge, err := cacheControlHeader.MaxAge()
	if err != nil {
		return c.next.Freshness(ctx, header, cacheControlHeader)
	}

	date, err := dateFromHeader(header)
	if err != nil {
		return c.next.Freshness(ctx, header, cacheControlHeader)
	}

	return freshnessFromMaxAge(maxAge, date, c.clock), nil
}

type expireFreshnessChecker struct {
	next freshnessChecker
}

func (c expireFreshnessChecker) Freshness(ctx context.Context, header http.Header, cacheControlHeader CacheControl) (Freshness, error) {
	expires, err := expiresFromHeader(header)
	if err != nil {
		return c.next.Freshness(ctx, header, cacheControlHeader)
	}

	date, err := dateFromHeader(header)
	if err != nil {
		return c.next.Freshness(ctx, header, cacheControlHeader)
	}

	return freshnessFromExpire(expires, date), nil
}

type ageFreshnessChecker struct {
	next freshnessChecker
}

func (c ageFreshnessChecker) Freshness(ctx context.Context, header http.Header, cacheControlHeader CacheControl) (Freshness, error) {

	maxAge, err := cacheControlHeader.MaxAge()
	if err != nil {
		return c.next.Freshness(ctx, header, cacheControlHeader)
	}

	age, err := ageFromHeader(header)
	if err != nil {
		return c.next.Freshness(ctx, header, cacheControlHeader)
	}

	return freshnessFromAge(age, maxAge), nil
}

type transparentFreshness struct{}

func (c transparentFreshness) Freshness(ctx context.Context, header http.Header, cacheControlHeader CacheControl) (Freshness, error) {
	return FreshnesTransparent, nil
}
