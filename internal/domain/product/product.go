package product

import (
	"time"
	"uuid"
)

type Product struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Slug          string    `json:"slug"`
	PriceMinor    int       `json:"price_minor"`
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	// DeletedAt is nil while the product is live. It is nullable in the schema,
	// so it must stay a pointer to scan a NULL without error.
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
