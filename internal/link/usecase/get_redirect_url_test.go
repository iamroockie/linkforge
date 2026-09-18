package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iamroockie/linkforge/internal/link"
	"github.com/iamroockie/linkforge/internal/link/linktest"
	"github.com/iamroockie/linkforge/internal/link/usecase"
	"github.com/iamroockie/linkforge/internal/platform/clock"
)

func TestGetRedirectURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := NewMockLinkProvider(ctrl)
	l := linktest.Link(t)
	provider.EXPECT().GetByAlias(t.Context(), "abc").Return(l, nil)

	uc := usecase.NewGetRedirectURL(provider)
	url, err := uc.Get(t.Context(), "abc")

	require.NoError(t, err)
	assert.Equal(t, l.URL, url)
}

func TestGetRedirectURL_ProviderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := NewMockLinkProvider(ctrl)
	providerErr := errors.New("test")
	provider.EXPECT().GetByAlias(t.Context(), "abc").Return(link.Link{}, providerErr)

	uc := usecase.NewGetRedirectURL(provider)
	url, err := uc.Get(t.Context(), "abc")

	require.ErrorIs(t, err, providerErr)
	assert.Empty(t, url)
}

func TestGetRedirectURL_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := NewMockLinkProvider(ctrl)
	provider.EXPECT().GetByAlias(t.Context(), "abc").Return(link.Link{}, link.ErrLinkNotFound)

	uc := usecase.NewGetRedirectURL(provider)
	url, err := uc.Get(t.Context(), "abc")

	require.ErrorIs(t, err, link.ErrLinkNotFound)
	assert.Empty(t, url)
}

func TestGetRedirectURL_LinkExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := NewMockLinkProvider(ctrl)
	l := linktest.Link(t)
	l.ExpiresAt = new(clock.Now().Add(-time.Minute))
	provider.EXPECT().GetByAlias(t.Context(), "abc").Return(l, nil)

	uc := usecase.NewGetRedirectURL(provider)
	url, err := uc.Get(t.Context(), "abc")

	require.ErrorIs(t, err, link.ErrLinkNotFound)
	assert.Empty(t, url)
}
