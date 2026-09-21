package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
	"github.com/iamroockie/linkforge/internal/platform/httpx/middleware"
)

func RateLimit(l Limiter, resolver httpx.ClientIPResolver, prefix string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return httpx.Handle(func(w http.ResponseWriter, r *http.Request) (*httpx.Response, error) {
			ip, err := resolver.Resolve(r)
			if err != nil {
				return nil, httpx.BadRequestError()
			}

			decision, err := l.Allow(r.Context(), prefix+":"+ip.String())
			if err != nil {
				return nil, httpx.InternalError(err)
			}

			if !decision.Allowed {
				seconds := decision.RetryAfter / time.Second
				if decision.RetryAfter%time.Second > 0 {
					seconds++
				}
				w.Header().Set("Retry-After", strconv.FormatInt(int64(max(1, seconds)), 10))
				return nil, httpx.TooManyRequestsError()
			}

			next.ServeHTTP(w, r)

			return nil, nil
		})
	}
}
