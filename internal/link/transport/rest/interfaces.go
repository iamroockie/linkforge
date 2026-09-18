//go:generate mockgen -source=interfaces.go -destination=mocks_test.go -package=rest_test
package rest

import (
	"context"

	"github.com/iamroockie/linkforge/internal/link"
)

type LinkCreator interface {
	Create(ctx context.Context, l link.CreateLinkParams) (link.Link, error)
}

type RedirectURLProvider interface {
	Get(ctx context.Context, alias string) (string, error)
}
