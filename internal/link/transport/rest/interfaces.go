//go:generate mockgen -source=interfaces.go -destination=mocks_test.go -package=rest_test
package rest

import (
	"context"

	"github.com/iamroockie/linkforge/internal/link"
)

type LinkCreator interface {
	Create(context.Context, link.CreateLinkParams) (link.Link, error)
}
