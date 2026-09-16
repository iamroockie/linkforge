package api_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iamroockie/linkforge/internal/app/api"
	"github.com/iamroockie/linkforge/internal/platform/pg/pgtest"
)

func TestNewApp(t *testing.T) {
	pool := pgtest.NewPool(t)

	app := api.NewApp(pool, slog.New(slog.DiscardHandler))

	assert.NotNil(t, app.PgxPool)
	assert.NotNil(t, app.Logger)
	assert.NotZero(t, app.Link.CreateLinkUC)
}
