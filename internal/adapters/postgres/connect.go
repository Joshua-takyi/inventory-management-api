package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshua-takyi/stockflow-api/internal/domain"
)

func InitDb(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, domain.ErrDbInitialization
	}
	return pool, err
}
