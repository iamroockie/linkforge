package middleware_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
	"github.com/iamroockie/linkforge/internal/platform/httpx/httpxtest"
	"github.com/iamroockie/linkforge/internal/platform/httpx/middleware"
)

type logRecord struct {
	level     slog.Level
	msg       string
	requestID string
	attrs     map[string]any
}

type recordHandler struct {
	records []logRecord
}

func (h *recordHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *recordHandler) Handle(ctx context.Context, rec slog.Record) error {
	attrs := make(map[string]any, rec.NumAttrs())
	rec.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	h.records = append(h.records, logRecord{
		level:     rec.Level,
		msg:       rec.Message,
		requestID: middleware.GetRequestID(ctx).String(),
		attrs:     attrs,
	})

	return nil
}

func (h *recordHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *recordHandler) WithGroup(string) slog.Handler      { return h }

func newRecordLogger() (*slog.Logger, *recordHandler) {
	h := &recordHandler{}

	return slog.New(h), h
}

func panicHandler() http.Handler {
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })
}

func errorHandler(err error) http.Handler {
	return httpx.Handle(func(http.ResponseWriter, *http.Request) (*httpx.Response, error) {
		return nil, err
	})
}

func noContentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}

func requireLoggedError(t *testing.T, rec logRecord, want string) {
	t.Helper()

	assert.Equal(t, slog.LevelError, rec.level)
	assert.Equal(t, "http request error", rec.msg)
	err, ok := rec.attrs["error"].(error)
	require.True(t, ok)
	require.ErrorContains(t, err, want)
}

func TestErrorLog(t *testing.T) {
	tests := map[string]struct {
		handler   http.Handler
		wantError string
	}{
		"panic": {
			handler:   middleware.Recover()(panicHandler()),
			wantError: "boom",
		},
		"server error": {
			handler:   errorHandler(errors.New("db is down")),
			wantError: "db is down",
		},
		"client error": {
			handler: errorHandler(httpx.BadRequestError()),
		},
		"no error": {
			handler: noContentHandler(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			log, logs := newRecordLogger()

			httpxtest.DoJSON(t, middleware.ErrorLog(log)(test.handler), http.MethodGet, "/", nil)

			if test.wantError == "" {
				assert.Empty(t, logs.records)
				return
			}
			require.Len(t, logs.records, 1)
			requireLoggedError(t, logs.records[0], test.wantError)
		})
	}
}

func TestErrorLog_ReusesTracker(t *testing.T) {
	log, logs := newRecordLogger()
	ctx := httpx.ReportErrorCtx(t.Context())
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

	middleware.ErrorLog(log)(errorHandler(errors.New("db is down"))).ServeHTTP(w, r)

	require.Len(t, logs.records, 1)
	require.ErrorContains(t, httpx.GetReportedError(ctx), "db is down")
}

func TestRequestLog(t *testing.T) {
	log, logs := newRecordLogger()
	h := middleware.RequestLog(log)(errorHandler(httpx.BadRequestError()))

	httpxtest.DoJSON(t, h, http.MethodPost, "/links", nil)

	require.Len(t, logs.records, 1)
	rec := logs.records[0]
	assert.Equal(t, slog.LevelInfo, rec.level)
	assert.Equal(t, "http request", rec.msg)
	assert.Equal(t, http.MethodPost, rec.attrs["method"])
	assert.Equal(t, "/links", rec.attrs["path"])
	assert.Equal(t, int64(http.StatusBadRequest), rec.attrs["status"])
	assert.Contains(t, rec.attrs, "duration_ms")
}

func TestRequestLog_Quiet(t *testing.T) {
	log, logs := newRecordLogger()
	h := middleware.RequestLog(log, "/healthz")(noContentHandler())

	httpxtest.DoJSON(t, h, http.MethodGet, "/healthz", nil)
	httpxtest.DoJSON(t, h, http.MethodGet, "/links", nil)

	require.Len(t, logs.records, 1)
	assert.Equal(t, "/links", logs.records[0].attrs["path"])
}
