package httpx_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
	"github.com/iamroockie/linkforge/internal/platform/httpx/httpxtest"
)

func TestHealthz(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)

	httpx.Healthz().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := httpxtest.DecodeJSON[httpx.Map](t, w)
	assert.Equal(t, "ok", resp["status"])
}

func TestReadyz_OK(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	checkers := []httpx.Check{
		func(_ context.Context) error { return nil },
		func(_ context.Context) error { return nil },
	}

	httpx.Readyz(2*time.Second, checkers...).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := httpxtest.DecodeJSON[httpx.Map](t, w)
	assert.Equal(t, "ready", resp["status"])
}

func TestReadyz_OneCheckerFail(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	checkers := []httpx.Check{
		func(_ context.Context) error { return nil },
		func(_ context.Context) error { return errors.New("one fail") },
	}

	httpx.Readyz(2*time.Second, checkers...).ServeHTTP(w, r)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	httpxtest.CheckResponseError(t, w, "service_unavailable", "Service unavailable")
}

func TestReadyz_AllCheckersFail(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	checkers := []httpx.Check{
		func(_ context.Context) error { return errors.New("one fail") },
		func(_ context.Context) error { return errors.New("two fail") },
	}

	httpx.Readyz(2*time.Second, checkers...).ServeHTTP(w, r)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	httpxtest.CheckResponseError(t, w, "service_unavailable", "Service unavailable")
}
