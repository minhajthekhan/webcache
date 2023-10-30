package webcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFreshnessFromMaxAge(t *testing.T) {
	ageInSeconds := 100
	responseDated := time.Now().Add(-3 * time.Minute)
	assert.Equal(t, FreshnessStale, freshnessFromMaxAge(ageInSeconds, responseDated))

	ageInSeconds = 100
	responseDated = time.Now().Add(-2 * time.Minute)
	assert.Equal(t, FreshnessStale, freshnessFromMaxAge(ageInSeconds, responseDated))

	ageInSeconds = 100
	responseDated = time.Now().Add(-1 * time.Minute)
	assert.Equal(t, FreshnessFresh, freshnessFromMaxAge(ageInSeconds, responseDated))
}

func TestFreshnessFromAge(t *testing.T) {
	assert.Equal(t, FreshnesTransparent, freshnessFromAge(100, 100, time.Now().Add(-3*time.Minute)))
	assert.Equal(t, FreshnessFresh, freshnessFromAge(10, 100, time.Now().Add(-1*time.Minute)))
	assert.Equal(t, FreshnessStale, freshnessFromAge(40, 100, time.Now().Add(-1*time.Minute)))

}

func TestFreshnessFromExpire(t *testing.T) {
	assert.Equal(t, FreshnessStale, freshnessFromExpire(time.Now(), time.Now().Add(5*time.Minute)))
	assert.Equal(t, FreshnessFresh, freshnessFromExpire(time.Now(), time.Now().Add(-5*time.Minute)))
	assert.Equal(t, FreshnesTransparent, freshnessFromExpire(time.Time{}, time.Now().Add(-5*time.Minute)))
}
