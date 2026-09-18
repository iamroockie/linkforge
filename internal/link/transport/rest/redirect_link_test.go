package rest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/iamroockie/linkforge/internal/link"
	"github.com/iamroockie/linkforge/internal/link/transport/rest"
	"github.com/iamroockie/linkforge/internal/platform/httpx/httpxtest"
)

func TestRedirectLink_OK(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	alias := "abc"
	r.SetPathValue("alias", alias)
	ctrl := gomock.NewController(t)
	prodiver := NewMockRedirectURLProvider(ctrl)
	prodiver.EXPECT().Get(gomock.Any(), alias).Return("http://test", nil)

	rest.RedirectLink(prodiver).ServeHTTP(w, r)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "http://test", w.Header().Get("Location"))
}

func TestRedirectLink_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	alias := "abc"
	r.SetPathValue("alias", alias)
	ctrl := gomock.NewController(t)
	prodiver := NewMockRedirectURLProvider(ctrl)
	prodiver.EXPECT().Get(gomock.Any(), alias).Return("", link.ErrLinkNotFound)

	rest.RedirectLink(prodiver).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, w.Header().Get("Location"))
	httpxtest.CheckResponseError(t, w, "link_not_found", "Link not found")
}
