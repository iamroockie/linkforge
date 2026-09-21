package ratelimit

import "time"

type Decision struct {
	Allowed    bool
	RetryAfter time.Duration
}
