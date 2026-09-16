package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/iamroockie/linkforge/internal/link"
)

type CreateLink struct {
	saver LinkSaver
}

func NewCreateLink(saver LinkSaver) CreateLink {
	return CreateLink{saver}
}

func (uc CreateLink) Create(ctx context.Context, p link.CreateLinkParams) (link.Link, error) {
	for range 5 {
		l, err := link.NewLink(p)
		if err != nil {
			return link.Link{}, fmt.Errorf("create link: %w", err)
		}

		err = uc.saver.Save(ctx, l)
		if err == nil {
			return l, nil
		}
		if !errors.Is(err, link.ErrAliasTaken) {
			return link.Link{}, fmt.Errorf("save link: %w", err)
		}
	}

	return link.Link{}, errors.New("create link: alias taken after 5 retries")
}
