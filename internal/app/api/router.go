package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	link "github.com/iamroockie/linkforge/internal/link/transport/rest"
	"github.com/iamroockie/linkforge/internal/platform/httpx"
	"github.com/iamroockie/linkforge/internal/platform/httpx/middleware"
)

func NewRouter(app App) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /healthz", httpx.Healthz())
	mux.Handle("GET /readyz", httpx.Readyz(2*time.Second, checkPostgres(app.PgxPool)))

	mux.Handle("POST /links", link.CreateLink(app.Link.CreateLinkUC))
	mux.Handle("GET /r/{alias}", link.RedirectLink(app.Link.GetRedirectURLUC))

	handler := httpx.RouteErrors(mux)
	handler = middleware.Recover()(handler)
	handler = middleware.ErrorLog(app.Logger)(handler)
	handler = middleware.RequestLog(app.Logger, "/healthz", "readyz")(handler)
	handler = middleware.RequestID()(handler)

	return handler
}

func checkPostgres(pool *pgxpool.Pool) httpx.Check {
	return func(ctx context.Context) error {
		if err := pool.Ping(ctx); err != nil {
			return fmt.Errorf("ping postgres: %w", err)
		}
		return nil
	}
}
