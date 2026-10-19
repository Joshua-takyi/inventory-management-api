package container

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshua-takyi/stockflow-api/config"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/postgres"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/postgres/migrations"
	"github.com/joshua-takyi/stockflow-api/internal/domain"
)

type Container struct {
	Config *config.Config
	logger *slog.Logger
}

func NewContainer(ctx context.Context) (*Container, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config %w", err)
	}

	pool, err := postgres.InitDb(ctx, cfg.DbUrl)
	if err != nil {
		return nil, fmt.Errorf("db init failure %w", err)
	}

	if err := migrations.Migrate(ctx, pool); err != nil {
		return nil, domain.ErrDbMigration
	}

	ctn := &Container{
		Config: cfg,
	}
	return ctn, nil
}
