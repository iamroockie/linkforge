package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
	"github.com/iamroockie/linkforge/internal/platform/httpx/httpxtest"
	"github.com/iamroockie/linkforge/internal/platform/httpx/middleware"
)

func TestRecover_Panic(t *testing.T) {
	ctx := httpx.ReportErrorCtx(t.Context())
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

	h := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })
	middleware.Recover()(h).ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	httpxtest.CheckResponseError(t, w, string(httpx.CodeInternalError), httpx.MsgInternalError)
	err := httpx.GetReportedError(r.Context())
	require.ErrorContains(t, err, "boom")
	require.ErrorContains(t, err, "panic recover")
	require.ErrorContains(t, err, "call stack")
}

func TestRecover_AbortHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	h := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) })

	require.PanicsWithValue(t, http.ErrAbortHandler, func() {
		middleware.Recover()(h).ServeHTTP(w, r)
	})
}

func TestRecover_PanicAfterWrite(t *testing.T) {
	ctx := httpx.ReportErrorCtx(t.Context())
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		panic("boom")
	})
	middleware.Recover()(h).ServeHTTP(w, r)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Empty(t, w.Body.String())
	require.ErrorContains(t, httpx.GetReportedError(r.Context()), "boom")
}

func TestRecover_WithoutPanic(t *testing.T) {
	ctx := httpx.ReportErrorCtx(t.Context())
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	middleware.Recover()(h).ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Empty(t, w.Body.String())
	require.NoError(t, httpx.GetReportedError(r.Context()))
}
