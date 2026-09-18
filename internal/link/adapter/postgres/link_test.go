package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/link"
	"github.com/iamroockie/linkforge/internal/link/adapter/postgres"
	"github.com/iamroockie/linkforge/internal/link/linktest"
	"github.com/iamroockie/linkforge/internal/platform/pg/pgtest"
)

func TestLinkRepository_Save(t *testing.T) {
	t.Parallel()
	pool := pgtest.NewPool(t)
	repo := postgres.NewLinkRepository(pool)
	l := linktest.Link(t)

	err := repo.Save(t.Context(), l)
	require.NoError(t, err)

	var (
		alias     string
		url       string
		expiresAt *time.Time
	)
	sql := `SELECT alias, url, expires_at FROM links WHERE alias = $1`
	err = pool.QueryRow(t.Context(), sql, l.Alias).Scan(&alias, &url, &expiresAt)
	require.NoError(t, err)
	assert.Equal(t, l.Alias, alias)
	assert.Equal(t, l.URL, url)
	require.NotNil(t, expiresAt)
	assert.WithinDuration(t, *l.ExpiresAt, *expiresAt, time.Microsecond)
}

func TestLinkRepository_Save_AliasTaken(t *testing.T) {
	t.Parallel()
	pool := pgtest.NewPool(t)
	repo := postgres.NewLinkRepository(pool)
	l := linktest.Link(t)

	err := repo.Save(t.Context(), l)
	require.NoError(t, err)

	err = repo.Save(t.Context(), l)
	require.ErrorIs(t, err, link.ErrAliasTaken)
}

func TestLinkRepository_Save_Error(t *testing.T) {
	t.Parallel()
	pool := pgtest.NewPool(t)
	repo := postgres.NewLinkRepository(pool)
	l := linktest.Link(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := repo.Save(ctx, l)

	require.ErrorIs(t, err, context.Canceled)
	require.ErrorContains(t, err, "insert link")
}

func TestLinkRepository_GetByAlias(t *testing.T) {
	t.Parallel()
	pool := pgtest.NewPool(t)
	repo := postgres.NewLinkRepository(pool)
	l := linktest.Link(t)

	err := repo.Save(t.Context(), l)
	require.NoError(t, err)

	link, err := repo.GetByAlias(t.Context(), l.Alias)
	require.NoError(t, err)
	assert.Equal(t, l.Alias, link.Alias)
	assert.Equal(t, l.URL, link.URL)
	require.NotNil(t, link.ExpiresAt)
	assert.WithinDuration(t, *l.ExpiresAt, *link.ExpiresAt, time.Microsecond)
}

func TestLinkRepository_GetByAlias_NotFound(t *testing.T) {
	t.Parallel()
	pool := pgtest.NewPool(t)
	repo := postgres.NewLinkRepository(pool)

	l, err := repo.GetByAlias(t.Context(), "not-found")

	require.ErrorIs(t, err, link.ErrLinkNotFound)
	assert.Zero(t, l)
}

func TestLinkRepository_GetByAlias_Error(t *testing.T) {
	t.Parallel()
	pool := pgtest.NewPool(t)
	repo := postgres.NewLinkRepository(pool)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	l, err := repo.GetByAlias(ctx, "abc")

	require.ErrorIs(t, err, context.Canceled)
	require.ErrorContains(t, err, "get link by alias")
	assert.Zero(t, l)
}
