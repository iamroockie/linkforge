package memory

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/iamroockie/linkforge/internal/ratelimit"
)

type bucket struct {
	tokens  float64
	updated time.Time
}

type TokenBucket struct {
	mu          sync.Mutex
	buckets     map[string]bucket
	rate        float64
	burst       float64
	now         func() time.Time
	lastCleanup time.Time
}

func NewTokenBucket(rate float64, burst int, now func() time.Time) (*TokenBucket, error) {
	if rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) {
		return nil, errors.New("rate must be positive and finite")
	}

	if burst <= 0 {
		return nil, errors.New("burst must be positive")
	}

	period := float64(time.Second) / rate
	if period < 1 || period*float64(burst) >= float64(math.MaxInt64) {
		return nil, errors.New("refill duration is outside time.Duration range")
	}

	if now == nil {
		return nil, errors.New("clock is required")
	}

	return &TokenBucket{
		buckets: make(map[string]bucket), rate: rate, burst: float64(burst), now: now,
		lastCleanup: now(),
	}, nil
}

func (l *TokenBucket) Allow(ctx context.Context, key string) (ratelimit.Decision, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return ratelimit.Decision{}, err
	}
	now := l.now()
	l.cleanup(now)
	b, ok := l.buckets[key]
	if !ok {
		b = bucket{tokens: l.burst, updated: now}
	}
	if now.After(b.updated) {
		b.tokens = l.refilled(b, now)
		b.updated = now
	}
	decision := ratelimit.Decision{Allowed: b.tokens >= 1}
	if decision.Allowed {
		b.tokens--
	} else {
		wait := math.Ceil((1 - b.tokens) / l.rate * float64(time.Second))
		decision.RetryAfter = time.Duration(wait)
		if now.Before(b.updated) {
			// Recovery cannot start until the clock catches up with the saved timestamp.
			delay := b.updated.Sub(now)
			decision.RetryAfter += min(delay, time.Duration(math.MaxInt64)-decision.RetryAfter)
		}
	}
	l.buckets[key] = b
	return decision, nil
}

func (l *TokenBucket) refilled(b bucket, now time.Time) float64 {
	return min(l.burst, b.tokens+max(0, now.Sub(b.updated).Seconds())*l.rate)
}

func (l *TokenBucket) cleanup(now time.Time) {
	if now.Sub(l.lastCleanup) < time.Minute {
		return
	}
	for key, b := range l.buckets {
		if l.refilled(b, now) >= l.burst {
			delete(l.buckets, key)
		}
	}
	l.lastCleanup = now
}
