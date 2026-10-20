package dto

import (
	"time"

	"uuid"

	"github.com/joshua-takyi/stockflow-api/internal/domain"
	"github.com/joshua-takyi/stockflow-api/internal/domain/product"
)

type ProductInputDTO struct {
	Name          string `json:"name" validate:"min=1,max=255"`
	Description   string `json:"description" validate:"required,max=2000"`
	PriceMinor    int    `json:"price_minor" validate:"required,gte=0"`
	StockQuantity int    `json:"stock_quantity" validate:"required,gte=0"`
}

type ProductResponseDTO struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description"`
	PriceMinor    int       `json:"price_minor"`
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ProductGetByIdDTO struct {
	ID string `json:"id"`
}

// ProductListQueryDTO binds the query string for GET /products.
type ProductListQueryDTO struct {
	Page     int    `form:"page" validate:"omitempty,min=1"`
	PerPage  int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Search   string `form:"search"`
	Slug     string `form:"slug"`
	MinPrice *int   `form:"min_price" validate:"omitempty,min=0"`
	MaxPrice *int   `form:"max_price" validate:"omitempty,min=0"`
	Sort     string `form:"sort" validate:"omitempty,oneof=newest oldest price_asc price_desc"`
}

// ProductListResponseDTO is the paginated envelope returned to clients.
type ProductListResponseDTO = domain.Paginated[ProductResponseDTO]

// ProductUpdateDTO binds a PATCH body. Every field is optional; pointers let an
// omitted field be told apart from one explicitly set to zero.
type ProductUpdateDTO struct {
	Name          *string `json:"name" validate:"omitempty,min=1,max=255"`
	Description   *string `json:"description" validate:"omitempty,max=2000"`
	PriceMinor    *int    `json:"price_minor" validate:"omitempty,gte=0"`
	StockQuantity *int    `json:"stock_quantity" validate:"omitempty,gte=0"`
}

// ToUpdate maps the request onto the domain update type. The service layer owns
// this translation so the repository never sees an HTTP shape.
func (u *ProductUpdateDTO) ToUpdate() product.Update {
	return product.Update{
		Name:          u.Name,
		Description:   u.Description,
		PriceMinor:    u.PriceMinor,
		StockQuantity: u.StockQuantity,
	}
}
