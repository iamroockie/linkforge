package api

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iamroockie/linkforge/internal/link/adapter/postgres"
	"github.com/iamroockie/linkforge/internal/link/usecase"
)

type LinkModule struct {
	CreateLinkUC usecase.CreateLink
}

func newLinkModule(pool *pgxpool.Pool) LinkModule {
	linkRepo := postgres.NewLinkRepository(pool)

	return LinkModule{
		CreateLinkUC: usecase.NewCreateLink(linkRepo),
	}
}
