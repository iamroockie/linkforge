package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := httpx.NewResponseWriter(w)

			defer func() {
				rvr := recover()
				if rvr == nil {
					return
				}
				if err, ok := rvr.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(rvr)
				}

				err := fmt.Errorf("panic recover: %v: call stack: %s", rvr, debug.Stack())
				httpx.ReportError(r.Context(), err)

				if ww.Wrote() {
					return
				}

				e := httpx.ErrorEnvelope{Error: httpx.InternalError(nil)}
				httpx.WriteJSON(ww, r, http.StatusInternalServerError, e)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
