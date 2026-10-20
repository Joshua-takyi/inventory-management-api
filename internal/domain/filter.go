package domain

type ProductSort string

const (
	SortNewest    ProductSort = "newest"
	SortOldest    ProductSort = "oldest"
	SortPriceAsc  ProductSort = "price_asc"
	SortPriceDesc ProductSort = "price_desc"
)

// Filter holds the searchable/sortable criteria for listing products.
type Filter struct {
	Search   string
	Slug     string
	MinPrice *int
	MaxPrice *int
	Sort     ProductSort
}

// Pagination holds the raw paging input plus the derived offset.
type Pagination struct {
	Page    int
	PerPage int
}

// Offset is the number of rows to skip for the current page.
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// DefaultPagination returns sensible bounds when the caller sends nothing.
func DefaultPagination() Pagination {
	return Pagination{Page: 1, PerPage: 20}
}

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Normalize clamps page/per_page into valid ranges.
func (p Pagination) Normalize() Pagination {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.PerPage < 1 {
		p.PerPage = DefaultPerPage
	}
	if p.PerPage > MaxPerPage {
		p.PerPage = MaxPerPage
	}
	return p
}

// Paginated is the generic envelope returned to clients.
type Paginated[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// NewPaginated builds the envelope and derives total pages from the row count.
func NewPaginated[T any](items []T, total int, p Pagination) Paginated[T] {
	if items == nil {
		items = []T{}
	}
	totalPages := 0
	if p.PerPage > 0 {
		totalPages = (total + p.PerPage - 1) / p.PerPage
	}
	return Paginated[T]{
		Items:      items,
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
