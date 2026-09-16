package api

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	PgxPool *pgxpool.Pool
	Logger  *slog.Logger
	Link    LinkModule
}

func NewApp(pool *pgxpool.Pool, log *slog.Logger) App {
	return App{
		PgxPool: pool,
		Logger:  log,

		Link: newLinkModule(pool),
	}
}
