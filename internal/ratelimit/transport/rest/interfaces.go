package rest

import (
	"context"

	"github.com/iamroockie/linkforge/internal/ratelimit"
)

// Limiter atomically checks and consumes one request for a key.
// Denial is a Decision, not an error. Implementations must support concurrent calls.
type Limiter interface {
	Allow(ctx context.Context, key string) (ratelimit.Decision, error)
}
