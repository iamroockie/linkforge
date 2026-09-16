package httpx_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

func TestHanlder_Response(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)

	h := func(_ http.ResponseWriter, _ *http.Request) (*httpx.Response, error) {
		return httpx.NewResponse(http.StatusCreated, httpx.Map{"id": "1234"}), nil
	}
	httpx.Handle(h).ServeHTTP(w, r)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t, `{"id":"1234"}`, w.Body.String())
}

func TestHanlder_ErrorWithReport(t *testing.T) {
	err := errors.New("test error")
	ctx := httpx.ReportErrorCtx(t.Context())
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

	h := func(_ http.ResponseWriter, _ *http.Request) (*httpx.Response, error) {
		return nil, err
	}
	httpx.Handle(h).ServeHTTP(w, r)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.ErrorIs(t, httpx.GetReportedError(r.Context()), err)
}

func TestHanlder_ErrorWithoutReport(t *testing.T) {
	ctx := httpx.ReportErrorCtx(t.Context())
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

	h := func(_ http.ResponseWriter, _ *http.Request) (*httpx.Response, error) {
		return nil, httpx.BadRequestError()
	}
	httpx.Handle(h).ServeHTTP(w, r)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.NoError(t, httpx.GetReportedError(r.Context()))
}

func TestHanlder_NoResponseNoError(t *testing.T) {
	ctx := httpx.ReportErrorCtx(t.Context())
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

	h := func(w http.ResponseWriter, _ *http.Request) (*httpx.Response, error) {
		w.WriteHeader(http.StatusCreated)
		return nil, nil
	}
	httpx.Handle(h).ServeHTTP(w, r)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, httpx.GetReportedError(r.Context()))
}
