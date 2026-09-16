package httpx_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
	"github.com/iamroockie/linkforge/internal/platform/httpx/httpxtest"
)

func TestWriteJSON_OK(t *testing.T) {
	body := httpx.Map{"status": "test"}
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)

	httpx.WriteJSON(w, r, http.StatusOK, body)

	resp := httpxtest.DecodeJSON[httpx.Map](t, w)
	assert.Equal(t, "test", resp["status"])
}

func TestWriteJSON_MarshalFail(t *testing.T) {
	ctx := httpx.ReportErrorCtx(t.Context())
	body := httpx.Map{"fn": func() {}}
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

	httpx.WriteJSON(w, r, http.StatusOK, body)

	httpxtest.CheckResponseError(t, w, "internal_error", "Internal server error")
	require.ErrorContains(t, httpx.GetReportedError(ctx), "marshal response")
}

func TestParseRequestJSON_InvalidRequest(t *testing.T) {
	tests := map[string]string{
		"broken body":     `{"url":"http`,
		"double body":     `{"url":"http://test"}{"url":"http://test"}`,
		"field duplicate": `{"url":"http://test","url":"http://test"}`,
		"unknown field":   `{"url":"http://test","lifetime":"24h"}`,
	}

	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			type want struct {
				URL string  `json:"url"`
				TTL *string `json:"ttl,omitempty"`
			}
			body := strings.NewReader(payload)
			r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", body)
			r.Header.Set(httpx.HeaderContentType, httpx.MIMEApplicationJSON)

			resp, err := httpx.ParseRequestJSON[want](r)

			require.Error(t, err)
			assert.Zero(t, resp)
			e, ok := errors.AsType[httpx.Error](err)
			require.True(t, ok)
			assert.Equal(t, httpx.CodeInvalidRequest, e.Code)
			assert.Equal(t, httpx.MsgBadRequest, e.Message)
		})
	}
}

func TestParseRequestJSON_InvalidMediaType(t *testing.T) {
	body := strings.NewReader(`{"url":"http://test"}`)
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", body)
	r.Header.Set(httpx.HeaderContentType, "text/plain")

	resp, err := httpx.ParseRequestJSON[httpx.Map](r)

	require.Error(t, err)
	assert.Zero(t, resp)
	e, ok := errors.AsType[httpx.Error](err)
	require.True(t, ok)
	assert.Equal(t, httpx.CodeUnsupportedMediaType, e.Code)
	assert.Equal(t, httpx.MsgUnsupportedMediaType, e.Message)
}

func TestParseRequestJSON_TooLargeBody(t *testing.T) {
	body := strings.NewReader(fmt.Sprintf(`{"url":"%s"}`, strings.Repeat("b", 1<<10+1)))
	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", body)
	r.Header.Set(httpx.HeaderContentType, httpx.MIMEApplicationJSON)
	r.Body = http.MaxBytesReader(httptest.NewRecorder(), r.Body, 1<<10)

	resp, err := httpx.ParseRequestJSON[httpx.Map](r)

	require.Error(t, err)
	assert.Zero(t, resp)
	e, ok := errors.AsType[httpx.Error](err)
	require.True(t, ok)
	assert.Equal(t, httpx.CodePayloadTooLarge, e.Code)
	assert.Equal(t, httpx.MsgRequestBodyTooLarge, e.Message)
}
