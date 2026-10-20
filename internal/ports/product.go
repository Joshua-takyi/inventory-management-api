package ports

import (
	"context"

	"github.com/joshua-takyi/stockflow-api/internal/domain"
	"github.com/joshua-takyi/stockflow-api/internal/domain/product"
)

type ProductRepository interface {
	Create(ctx context.Context, input *product.Product) (*product.Product, error)
	GetByID(ctx context.Context, id string) (*product.Product, error)
	ListProducts(ctx context.Context, filter domain.Filter, page domain.Pagination) ([]*product.Product, int, error)
	UpdateProduct(ctx context.Context, id string, update product.Update) (*product.Product, error)
	Delete(ctx context.Context, id string) (string, error)
}
