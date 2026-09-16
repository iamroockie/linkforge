package middleware

import (
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

func ErrorLog(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := httpx.ReportErrorCtx(r.Context())

			next.ServeHTTP(w, r.WithContext(ctx))

			if err := httpx.GetReportedError(ctx); err != nil {
				log.ErrorContext(r.Context(), "http request error", "error", err)
			}
		})
	}
}

func RequestLog(log *slog.Logger, quiet ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := httpx.NewResponseWriter(w)

			start := time.Now()
			next.ServeHTTP(ww, r)
			elapsed := time.Since(start)

			if slices.Contains(quiet, r.URL.Path) {
				return
			}

			log.InfoContext(r.Context(), "http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", float64(elapsed.Microseconds())/1000,
			)
		})
	}
}
