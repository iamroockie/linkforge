package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iamroockie/linkforge/internal/link"
)

type LinkRepository struct {
	pool *pgxpool.Pool
}

func NewLinkRepository(pool *pgxpool.Pool) LinkRepository {
	return LinkRepository{pool}
}

func (r LinkRepository) Save(ctx context.Context, l link.Link) error {
	sql := `
		INSERT INTO links (alias, url, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (alias) DO NOTHING
	`
	tag, err := r.pool.Exec(ctx, sql, l.Alias, l.URL, l.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert link: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return link.ErrAliasTaken
	}

	return nil
}

func (r LinkRepository) GetByAlias(ctx context.Context, alias string) (link.Link, error) {
	sql := `SELECT alias, url, expires_at FROM links WHERE alias = $1`
	row := r.pool.QueryRow(ctx, sql, alias)

	var l link.Link
	err := row.Scan(&l.Alias, &l.URL, &l.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return link.Link{}, link.ErrLinkNotFound
		}
		return link.Link{}, fmt.Errorf("get link by alias: %w", err)
	}

	return l, nil
}
