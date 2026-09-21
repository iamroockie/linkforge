package memory

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenBucket(t *testing.T) {
	now := time.Now()
	limiter, err := NewTokenBucket(2, 2, func() time.Time { return now })
	require.NoError(t, err)
	for range 2 {
		d, err := limiter.Allow(t.Context(), "a")
		require.NoError(t, err)
		assert.True(t, d.Allowed)
		assert.Zero(t, d.RetryAfter)
	}
	for range 2 {
		d, err := limiter.Allow(t.Context(), "a")
		require.NoError(t, err)
		assert.False(t, d.Allowed)
		assert.Equal(t, 500*time.Millisecond, d.RetryAfter)
	}
	now = now.Add(250 * time.Millisecond)
	d, err := limiter.Allow(t.Context(), "a")
	require.NoError(t, err)
	assert.False(t, d.Allowed)
	assert.Equal(t, 250*time.Millisecond, d.RetryAfter)
	now = now.Add(250 * time.Millisecond)
	d, err = limiter.Allow(t.Context(), "a")
	require.NoError(t, err)
	assert.True(t, d.Allowed)
	d, err = limiter.Allow(t.Context(), "b")
	require.NoError(t, err)
	assert.True(t, d.Allowed)
	now = now.Add(time.Hour)
	for range 2 {
		d, err = limiter.Allow(t.Context(), "a")
		require.NoError(t, err)
		assert.True(t, d.Allowed)
	}
	d, err = limiter.Allow(t.Context(), "a")
	require.NoError(t, err)
	assert.False(t, d.Allowed)
	assert.NotContains(t, limiter.buckets, "b")
}

func TestTokenBucketCleanupPreservesActiveBucket(t *testing.T) {
	now := time.Now()
	limiter, err := NewTokenBucket(0.01, 1, func() time.Time { return now })
	require.NoError(t, err)
	_, err = limiter.Allow(t.Context(), "a")
	require.NoError(t, err)
	now = now.Add(time.Minute)
	_, err = limiter.Allow(t.Context(), "b")
	require.NoError(t, err)
	assert.Contains(t, limiter.buckets, "a")
	d, err := limiter.Allow(t.Context(), "a")
	require.NoError(t, err)
	assert.False(t, d.Allowed)
	assert.InDelta(t, float64(40*time.Second), float64(d.RetryAfter), 1)
}

func TestTokenBucketBackwardClock(t *testing.T) {
	now := time.Now()
	limiter, err := NewTokenBucket(1, 1, func() time.Time { return now })
	require.NoError(t, err)
	_, err = limiter.Allow(t.Context(), "a")
	require.NoError(t, err)
	now = now.Add(-time.Second)
	d, err := limiter.Allow(t.Context(), "a")
	require.NoError(t, err)
	assert.False(t, d.Allowed)
	assert.Equal(t, 2*time.Second, d.RetryAfter)
	now = now.Add(time.Second)
	d, err = limiter.Allow(t.Context(), "a")
	require.NoError(t, err)
	assert.False(t, d.Allowed)
	assert.Equal(t, time.Second, d.RetryAfter)
}

func TestTokenBucketConcurrent(t *testing.T) {
	now := time.Now()
	limiter, err := NewTokenBucket(1, 10, func() time.Time { return now })
	require.NoError(t, err)
	var allowed atomic.Int64
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			d, err := limiter.Allow(t.Context(), "a")
			if err != nil {
				t.Error(err)
				return
			}
			if d.Allowed {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()
	assert.EqualValues(t, 10, allowed.Load())
}

func TestTokenBucketCanceled(t *testing.T) {
	limiter, err := NewTokenBucket(1, 1, time.Now)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = limiter.Allow(ctx, "a")
	require.ErrorIs(t, err, context.Canceled)
	d, err := limiter.Allow(t.Context(), "a")
	require.NoError(t, err)
	assert.True(t, d.Allowed)
}

func TestTokenBucketValidation(t *testing.T) {
	for _, rate := range []float64{0, -1, math.NaN(), math.Inf(1), 1e-20, 1e20} {
		_, err := NewTokenBucket(rate, 1, time.Now)
		require.Error(t, err)
	}
	_, err := NewTokenBucket(1, 0, time.Now)
	require.Error(t, err)
	_, err = NewTokenBucket(1, 1, nil)
	require.Error(t, err)
}
