package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshua-takyi/stockflow-api/internal/domain"
	"github.com/joshua-takyi/stockflow-api/internal/domain/product"
)

type Repository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewRepository(p *pgxpool.Pool, l *slog.Logger) *Repository {
	return &Repository{
		pool:   p,
		logger: l,
	}
}

const productColumns = `id , name , slug ,description,price_minor, stock_quantity, created_at, updated_at`

func (r *Repository) Create(ctx context.Context, input *product.Product) (*product.Product, error) {
	if r.pool == nil {
		return nil, domain.ErrNoPgxConnection
	}
	query := `insert into products(name ,slug, description,price_minor, stock_quantity) values($1,$2,$3,$4,$5) returning ` + productColumns

	var product product.Product
	if err := r.pool.QueryRow(ctx, query, input.Name, input.Slug, input.Description, input.PriceMinor, input.StockQuantity).Scan(&product.ID, &product.Name, &product.Slug, &product.Description, &product.PriceMinor, &product.StockQuantity, &product.CreatedAt, &product.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return nil, domain.ErrConflict
			case "23514":
				return nil, domain.ErrCheckViolation
			default:
				r.logger.ErrorContext(ctx, "create product", "error", err)
				return nil, fmt.Errorf("create product %w", err)
			}
		}
		r.logger.ErrorContext(ctx, "create product", "error", err)
		return nil, fmt.Errorf("create product %w", err)
	}

	return &product, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*product.Product, error) {
	if r.pool == nil {
		return nil, domain.ErrNoPgxConnection
	}

	query := `select ` + productColumns + ` from products where id = $1 and ` + notDeleted

	product, err := ScanRow(r.pool.QueryRow(ctx, query, id))
	var pgConn *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgConn) {
			if pgConn.Code == "22P02" {
				return nil, domain.ErrInvalidUUID
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "get product by id", "error", err)
		return nil, fmt.Errorf("get product %w", err)
	}

	return product, nil
}

func (r *Repository) ListProducts(ctx context.Context, filter domain.Filter, page domain.Pagination) ([]*product.Product, int, error) {
	if r.pool == nil {
		return nil, 0, domain.ErrNoPgxConnection
	}

	b := &queryBuilder{}
	addFilters(b, filter)
	where := b.where()

	// Run count + page in one read-only tx so total matches the returned rows.
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		r.logger.ErrorContext(ctx, "begin list products", "error", err)
		return nil, 0, fmt.Errorf("list products %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var total int
	if err := tx.QueryRow(ctx, `select count(*) from products`+where, b.args...).Scan(&total); err != nil {
		r.logger.ErrorContext(ctx, "count products", "error", err)
		return nil, 0, fmt.Errorf("count products %w", err)
	}

	// Short-circuit: no rows means skip the page query entirely.
	if total == 0 {
		return []*product.Product{}, 0, nil
	}

	args := append(b.args[:len(b.args):len(b.args)], page.PerPage, page.Offset())
	query := `select ` + productColumns + ` from products` + where +
		orderBy(filter.Sort) +
		fmt.Sprintf(` limit $%d offset $%d`, len(args)-1, len(args))

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "list products", "error", err)
		return nil, 0, fmt.Errorf("list products %w", err)
	}
	defer rows.Close()

	products := make([]*product.Product, 0, page.PerPage)
	for rows.Next() {
		p, err := ScanRow(rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "scan product", "error", err)
			return nil, 0, fmt.Errorf("scan product %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "iterate products", "error", err)
		return nil, 0, fmt.Errorf("iterate products %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.ErrorContext(ctx, "commit list products", "error", err)
		return nil, 0, fmt.Errorf("list products %w", err)
	}

	return products, total, nil
}

// UpdateProduct applies a partial update and returns the stored row. Only the
// fields set on the update are written, so concurrent edits to different
// columns don't clobber each other.
func (r *Repository) UpdateProduct(ctx context.Context, id string, update product.Update) (*product.Product, error) {
	if r.pool == nil {
		return nil, domain.ErrNoPgxConnection
	}
	if update.IsEmpty() {
		return nil, domain.ErrNoFieldsToUpdate
	}

	query, args := buildProductUpdate(id, update)

	updated, err := ScanRow(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "22P02":
				return nil, domain.ErrInvalidUUID
			case "23505":
				return nil, domain.ErrConflict
			case "23514":
				return nil, domain.ErrCheckViolation
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "update product", "error", err)
		return nil, fmt.Errorf("update product %w", err)
	}

	return updated, nil
}

// Delete soft-deletes a product. The guard makes a second delete a not-found
// rather than a silent success, and RETURNING gives QueryRow a row to return.
func (r *Repository) Delete(ctx context.Context, id string) (string, error) {
	if r.pool == nil {
		return "", domain.ErrNoPgxConnection
	}

	query := `update products set deleted_at = now() where id = $1 and ` + notDeleted + ` returning id`

	var deletedID string
	err := r.pool.QueryRow(ctx, query, id).Scan(&deletedID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "22P02" {
			return "", domain.ErrInvalidUUID
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "delete product", "error", err)
		return "", fmt.Errorf("delete product %w", err)
	}

	return "product is deleted", nil
}

type Scan interface {
	Scan(dest ...any) error
}

func ScanRow(row Scan) (*product.Product, error) {
	product := &product.Product{}
	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.PriceMinor,
		&product.StockQuantity,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return product, err
}
