package container

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshua-takyi/stockflow-api/config"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/http/handlers"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/postgres"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/postgres/migrations"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/postgres/repository"
	"github.com/joshua-takyi/stockflow-api/internal/domain"
	"github.com/joshua-takyi/stockflow-api/internal/services"
)

type Container struct {
	Config         *config.Config
	Repository     *repository.Repository
	ProductService *services.ProductService
	ProductHandler *handlers.ProductHandler
	HealthHandler  *handlers.HealthHandler
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
		return nil, fmt.Errorf("%w: %v", domain.ErrDbMigration, err)
	}
	newRepository := repository.NewRepository(pool, Logger())

	newProductService := services.NewProductService(newRepository)
	newProductHandler := handlers.NewProductHandler(newProductService, Logger())
	newHealthHandler := handlers.NewHealthHandler(pool)
	ctn := &Container{
		Config:         cfg,
		Repository:     newRepository,
		ProductService: newProductService,
		ProductHandler: newProductHandler,
		HealthHandler:  newHealthHandler,
	}
	return ctn, nil
}

func Logger() *slog.Logger {
	return slog.Default()
}
