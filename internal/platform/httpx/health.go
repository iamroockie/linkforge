package httpx

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

type Check func(ctx context.Context) error

func Healthz() http.Handler {
	return Handle(func(_ http.ResponseWriter, _ *http.Request) (*Response, error) {
		return NewResponse(http.StatusOK, Map{"status": "ok"}), nil
	})
}

func Readyz(timeout time.Duration, checks ...Check) http.HandlerFunc {
	return Handle(func(_ http.ResponseWriter, r *http.Request) (*Response, error) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		errs := make([]error, len(checks))

		var wg sync.WaitGroup
		for i, check := range checks {
			wg.Go(func() { errs[i] = check(ctx) })
		}

		wg.Wait()

		if err := errors.Join(errs...); err != nil {
			return nil, NewError(
				http.StatusServiceUnavailable,
				CodeServiceUnavailable,
				MsgServiceUnavailable,
				err,
			)
		}

		return NewResponse(http.StatusOK, Map{"status": "ready"}), nil
	})
}
