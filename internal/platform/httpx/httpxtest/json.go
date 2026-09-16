package httpxtest

import (
	"bytes"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

func DoJSON(
	t *testing.T,
	h http.Handler,
	method string,
	path string,
	payload any,
) *httptest.ResponseRecorder {
	t.Helper()

	var body io.Reader
	if p, ok := payload.(io.Reader); ok {
		body = p
	} else if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal JSON: %v", err)
		}
		body = bytes.NewReader(b)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), method, path, body)
	r.Header.Set(httpx.HeaderContentType, httpx.MIMEApplicationJSON)

	h.ServeHTTP(w, r)

	return w
}

func DecodeJSON[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()

	mediaType := rec.Header().Get(httpx.HeaderContentType)
	if mediaType != httpx.MIMEApplicationJSON {
		t.Fatalf("expected content type %q, got %q", httpx.MIMEApplicationJSON, mediaType)
	}

	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v, json.RejectUnknownMembers(true)); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}

	return v
}

func CheckResponseError(t *testing.T, rec *httptest.ResponseRecorder, code, message string) {
	t.Helper()

	resp := DecodeJSON[httpx.Map](t, rec)
	err, ok := resp["error"].(map[string]any)
	require.True(t, ok, "error field is not a map")
	assert.Equal(t, code, err["code"])
	assert.Equal(t, message, err["message"])
}
