package webcache

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFreshnessFromMaxAge(t *testing.T) {

	ageInSeconds := 100
	responseDated := time.Now().Add(-3 * time.Minute)
	assert.Equal(t, FreshnessStale, freshnessFromMaxAge(ageInSeconds, responseDated, NewClock()))

	ageInSeconds = 100
	responseDated = time.Now().Add(-2 * time.Minute)
	assert.Equal(t, FreshnessStale, freshnessFromMaxAge(ageInSeconds, responseDated, NewClock()))

	ageInSeconds = 100
	responseDated = time.Now().Add(-1 * time.Minute)
	assert.Equal(t, FreshnessFresh, freshnessFromMaxAge(ageInSeconds, responseDated, NewClock()))

	ageInSeconds = 0
	responseDated = time.Now().Add(-1 * time.Minute)
	assert.Equal(t, FreshnessStale, freshnessFromMaxAge(ageInSeconds, responseDated, NewClock()))
}

func TestFreshnessFromAge(t *testing.T) {
	assert.Equal(t, FreshnessStale, freshnessFromAge(100, 100))
	assert.Equal(t, FreshnessFresh, freshnessFromAge(10, 100))
	assert.Equal(t, FreshnessStale, freshnessFromAge(120, 100))

}

func TestFreshnessFromExpire(t *testing.T) {
	assert.Equal(t, FreshnessStale, freshnessFromExpire(time.Now(), time.Now().Add(5*time.Minute)))
	assert.Equal(t, FreshnessFresh, freshnessFromExpire(time.Now(), time.Now().Add(-5*time.Minute)))
	assert.Equal(t, FreshnesTransparent, freshnessFromExpire(time.Time{}, time.Now().Add(-5*time.Minute)))
}

func TestFreshness(t *testing.T) {
	ctx := context.Background()
	headers := make(http.Header)
	headers.Add("Cache-Control", "max-age=120")
	headers.Add("Date", time.Now().Add(-1*time.Minute).Format(time.RFC850))
	cacheControl := newCacheControl(headers)
	checker := newFreshnerChecker(NewClock())
	freshness, err := checker.Freshness(ctx, headers, cacheControl)
	assert.NoError(t, err)
	assert.Equal(t, FreshnessFresh, freshness)

	headers = make(http.Header)
	headers.Add("Cache-Control", "max-age=40")
	headers.Add("Date", time.Now().Add(-1*time.Minute).Format(time.RFC850))
	cacheControl = newCacheControl(headers)
	checker = newFreshnerChecker(NewClock())
	freshness, err = checker.Freshness(ctx, headers, cacheControl)
	assert.NoError(t, err)
	assert.Equal(t, FreshnessStale, freshness)

	headers = make(http.Header)
	headers.Add("Date", time.Now().Add(-1*time.Minute).Format(time.RFC850))
	cacheControl = newCacheControl(headers)
	checker = newFreshnerChecker(NewClock())
	freshness, err = checker.Freshness(ctx, headers, cacheControl)
	assert.NoError(t, err)
	assert.Equal(t, FreshnesTransparent, freshness)

	headers = make(http.Header)
	headers.Add("Date", time.Now().Add(-1*time.Minute).Format(time.RFC850))
	headers.Add("Cache-Control", "max-age=1000, no-cache")
	cacheControl = newCacheControl(headers)
	checker = newFreshnerChecker(NewClock())
	freshness, err = checker.Freshness(ctx, headers, cacheControl)
	assert.NoError(t, err)
	assert.Equal(t, FreshnesTransparent, freshness)

	headers = make(http.Header)
	headers.Add("Date", time.Now().Add(-1*time.Minute).Format(time.RFC850))
	headers.Add("Cache-Control", "max-age=0, must-revalidate")
	cacheControl = newCacheControl(headers)
	checker = newFreshnerChecker(NewClock())
	freshness, err = checker.Freshness(ctx, headers, cacheControl)
	assert.NoError(t, err)
	assert.Equal(t, FreshnesTransparent, freshness)

	headers = make(http.Header)
	headers.Add("Date", time.Now().Add(-1*time.Minute).Format(time.RFC850))
	headers.Add("Cache-Control", "max-age=100, must-revalidate")
	cacheControl = newCacheControl(headers)
	checker = newFreshnerChecker(NewClock())
	freshness, err = checker.Freshness(ctx, headers, cacheControl)
	assert.NoError(t, err)
	assert.Equal(t, FreshnessFresh, freshness)

}
