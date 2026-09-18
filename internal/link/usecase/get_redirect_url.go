package usecase

import (
	"context"
	"fmt"

	"github.com/iamroockie/linkforge/internal/link"
	"github.com/iamroockie/linkforge/internal/platform/clock"
)

type GetRedirectURL struct {
	provider LinkProvider
}

func NewGetRedirectURL(provider LinkProvider) GetRedirectURL {
	return GetRedirectURL{provider}
}

func (uc GetRedirectURL) Get(ctx context.Context, alias string) (string, error) {
	l, err := uc.provider.GetByAlias(ctx, alias)
	if err != nil {
		return "", fmt.Errorf("get link: %w", err)
	}

	if l.ExpiresAt != nil && l.ExpiresAt.Before(clock.Now()) {
		return "", fmt.Errorf("link expired: %w", link.ErrLinkNotFound)
	}

	return l.URL, nil
}
