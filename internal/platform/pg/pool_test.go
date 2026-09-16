package pg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/platform/pg"
	"github.com/iamroockie/linkforge/internal/platform/pg/pgtest"
)

func TestNewPool(t *testing.T) {
	base := pgtest.NewPool(t)

	pool, err := pg.NewPool(t.Context(), base.Config().ConnString())

	require.NoError(t, err)
	require.NotNil(t, pool)
	assert.Equal(t, int32(2), pool.Config().MinIdleConns)

	pool.Close()
}

func TestNewPool_InvalidConnString(t *testing.T) {
	pool, err := pg.NewPool(t.Context(), "postgres//")

	require.ErrorContains(t, err, "parse connection string")
	assert.Nil(t, pool)
}
