//go:generate mockgen -source=interfaces.go -destination=mocks_test.go -package=usecase_test
package usecase

import (
	"context"

	"github.com/iamroockie/linkforge/internal/link"
)

type LinkSaver interface {
	Save(context.Context, link.Link) error
}
