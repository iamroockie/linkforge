package rest_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iamroockie/linkforge/internal/link"
	"github.com/iamroockie/linkforge/internal/link/transport/rest"
	"github.com/iamroockie/linkforge/internal/platform/clock"
	"github.com/iamroockie/linkforge/internal/platform/httpx"
	"github.com/iamroockie/linkforge/internal/platform/httpx/httpxtest"
)

func TestCreateLink_OK_WithTTL(t *testing.T) {
	url := "http://test"
	ttl := 24 * time.Hour
	body := httpx.Map{"url": url, "ttl": ttl.String()}
	createParams := link.CreateLinkParams{URL: url, TTL: &ttl}
	l := link.Link{Alias: "abc", URL: url, ExpiresAt: new(clock.Now().Add(ttl))}
	ctrl := gomock.NewController(t)
	creator := NewMockLinkCreator(ctrl)
	creator.EXPECT().Create(t.Context(), createParams).Return(l, nil)

	handler := rest.CreateLink(creator)
	rec := httpxtest.DoJSON(t, handler, http.MethodPost, "/", body)

	require.Equal(t, http.StatusCreated, rec.Code)
	resp := httpxtest.DecodeJSON[httpx.Map](t, rec)
	assert.Equal(t, l.Alias, resp["alias"])
	assert.Equal(t, l.URL, resp["url"])
	assert.Equal(t, l.ExpiresAt.Format("2006-01-02T15:04:05.999999999Z"), resp["expires_at"])
}

func TestCreateLink_OK_WithoutTTL(t *testing.T) {
	url := "http://test"
	body := httpx.Map{"url": url}
	l := link.Link{Alias: "abc", URL: url}
	createParams := link.CreateLinkParams{URL: url}
	ctrl := gomock.NewController(t)
	creator := NewMockLinkCreator(ctrl)
	creator.EXPECT().Create(t.Context(), createParams).Return(l, nil)

	handler := rest.CreateLink(creator)
	rec := httpxtest.DoJSON(t, handler, http.MethodPost, "/", body)

	require.Equal(t, http.StatusCreated, rec.Code)
	resp := httpxtest.DecodeJSON[httpx.Map](t, rec)
	assert.Equal(t, l.Alias, resp["alias"])
	assert.Equal(t, l.URL, resp["url"])
	assert.Nil(t, resp["expires_at"])
}

func TestCreateLink_InvalidRequest(t *testing.T) {
	tests := map[string]string{
		"empty ttl":        `{"url":"http://test","ttl":""}`,
		"invalid ttl":      `{"url":"http://test","ttl":"abc"}`,
		"invalid type ttl": `{"url":"http://test","ttl":100}`,
		"invalid type url": `{"url":false,"ttl":"24h"}`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			creator := NewMockLinkCreator(ctrl)
			creator.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

			handler := rest.CreateLink(creator)
			rec := httpxtest.DoJSON(t, handler, http.MethodPost, "/", strings.NewReader(body))

			require.Equal(t, http.StatusBadRequest, rec.Code)
			httpxtest.CheckResponseError(t, rec, "invalid_request", "Invalid request")
		})
	}
}

func TestCreateLink_InvalidMediaType(t *testing.T) {
	body := strings.NewReader(`{"url":"http://test"}`)
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", body)
	r.Header.Set(httpx.HeaderContentType, "text/plain")
	ctrl := gomock.NewController(t)
	creator := NewMockLinkCreator(ctrl)
	creator.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	handler := rest.CreateLink(creator)
	handler.ServeHTTP(w, r)

	require.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	httpxtest.CheckResponseError(t, w, "unsupported_media_type", "Unsupported media type")
}

func TestCreateLink_TooLargeBody(t *testing.T) {
	body := strings.NewReader(fmt.Sprintf(`{"url":"%s"}`, strings.Repeat("b", 1<<11)))
	ctrl := gomock.NewController(t)
	creator := NewMockLinkCreator(ctrl)
	creator.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	handler := rest.CreateLink(creator)
	rec := httpxtest.DoJSON(t, handler, http.MethodPost, "/", body)

	require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	httpxtest.CheckResponseError(t, rec, "payload_too_large", "Request body too large")
}

func TestCreateLink_ValidationError_Single(t *testing.T) {
	body := strings.NewReader(`{"url":"test"}`)
	ctrl := gomock.NewController(t)
	creator := NewMockLinkCreator(ctrl)
	creator.EXPECT().Create(gomock.Any(), gomock.Any()).Return(link.Link{}, link.ErrURLMalformed)

	handler := rest.CreateLink(creator)
	rec := httpxtest.DoJSON(t, handler, http.MethodPost, "/", body)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	resp := httpxtest.DecodeJSON[httpx.ErrorEnvelope](t, rec)
	assert.Equal(t, httpx.MsgValidationError, resp.Error.Message)
	assert.Equal(t, httpx.CodeValidationFailed, resp.Error.Code)
	require.Len(t, resp.Error.Details, 1)
	assert.Equal(t, "url", resp.Error.Details[0].Field)
	assert.Equal(t, httpx.FieldCodeInvalidFormat, resp.Error.Details[0].Code)
}

func TestCreateLink_ValidationError_Multiple(t *testing.T) {
	tests := map[string]struct {
		body          string
		creatorErr    error
		wantFieldErrs []httpx.FieldError
	}{
		"url malformed": {
			body:       `{"url":""}`,
			creatorErr: link.ErrURLRequired,
			wantFieldErrs: []httpx.FieldError{
				{Field: "url", Code: httpx.FieldCodeRequired},
			},
		},
		"url malformed and ttl negative": {
			body:       `{"url":"ftp://test","ttl":"-1h"}`,
			creatorErr: errors.Join(link.ErrURLUnsupportedScheme, link.ErrTTLNonPositive),
			wantFieldErrs: []httpx.FieldError{
				{Field: "url", Code: rest.FieldCodeUnsupportedScheme},
				{Field: "ttl", Code: httpx.FieldCodeNonPositive},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			creator := NewMockLinkCreator(ctrl)
			creator.EXPECT().Create(gomock.Any(), gomock.Any()).Return(link.Link{}, test.creatorErr)

			handler := rest.CreateLink(creator)
			rec := httpxtest.DoJSON(t, handler, http.MethodPost, "/", strings.NewReader(test.body))

			require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			resp := httpxtest.DecodeJSON[httpx.ErrorEnvelope](t, rec)
			assert.Equal(t, httpx.MsgValidationError, resp.Error.Message)
			assert.Equal(t, httpx.CodeValidationFailed, resp.Error.Code)
			require.ElementsMatch(t, test.wantFieldErrs, resp.Error.Details)
		})
	}
}
