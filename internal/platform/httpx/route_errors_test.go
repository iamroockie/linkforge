package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
	"github.com/iamroockie/linkforge/internal/platform/httpx/httpxtest"
)

func TestRouteErrors(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	tests := map[string]struct {
		route      string
		method     string
		wantStatus int
		assert     func(*testing.T, *httptest.ResponseRecorder)
	}{
		"found get": {
			route:  "/test",
			method: http.MethodGet,
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNoContent, w.Code)
			},
		},
		"found post": {
			route:  "/test",
			method: http.MethodPost,
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, w.Code)
			},
		},
		"method not allowed": {
			route:  "/test",
			method: http.MethodPut,
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
				assert.Equal(t, "GET, HEAD, POST", w.Header().Get("Allow"))
				resp := httpxtest.DecodeJSON[httpx.ErrorEnvelope](t, w)
				assert.Equal(t, httpx.CodeMethodNotAllowed, resp.Error.Code)
				assert.Equal(t, httpx.MsgMethodNotAllowed, resp.Error.Message)
			},
		},
		"not found": {
			route:  "/not-found",
			method: http.MethodGet,
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, w.Code)
				resp := httpxtest.DecodeJSON[httpx.ErrorEnvelope](t, w)
				assert.Equal(t, httpx.CodeRouteNotFound, resp.Error.Code)
				assert.Equal(t, httpx.MsgRouteNotFound, resp.Error.Message)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequestWithContext(t.Context(), test.method, test.route, nil)

			httpx.RouteErrors(mux).ServeHTTP(w, r)

			test.assert(t, w)
		})
	}
}
