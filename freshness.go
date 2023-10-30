package webcache

import (
	"time"
)

type Freshness int

const (
	FreshnessFresh Freshness = iota
	FreshnessStale
	FreshnesTransparent
)

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

func freshnessFromExpire(expireTime time.Time, responseDated time.Time) Freshness {
	if expireTime.IsZero() {
		return FreshnesTransparent
	}
	if expireTime.Before(responseDated) {
		return FreshnessStale
	}
	return FreshnessFresh
}

func freshnessFromAge(age int, maxAge int, responseDated time.Time) Freshness {
	return freshnessFromMaxAge(maxAge-age, responseDated)
}
