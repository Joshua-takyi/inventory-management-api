package ports

import (
	"context"

	"github.com/joshua-takyi/stockflow-api/internal/domain/product"
)

type ProductRepository interface {
	Create(ctx context.Context, input *product.Product) (*product.Product, error)
}
