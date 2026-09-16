package middleware_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iamroockie/linkforge/internal/platform/httpx/httpxtest"
	"github.com/iamroockie/linkforge/internal/platform/httpx/middleware"
)

func TestRequestID(t *testing.T) {
	h := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		assert.NotZero(t, middleware.GetRequestID(r.Context()))
	})

	rec := httpxtest.DoJSON(t, middleware.RequestID()(h), http.MethodGet, "/", nil)

	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}
