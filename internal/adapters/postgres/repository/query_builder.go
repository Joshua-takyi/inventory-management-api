package repository

import (
	"fmt"
	"strings"

	"github.com/joshua-takyi/stockflow-api/internal/domain"
)

// queryBuilder accumulates positional args and emits $N placeholders so the
// same WHERE clause can be shared by a COUNT and a SELECT without re-numbering.
type queryBuilder struct {
	conditions []string
	args       []any
}

// arg appends a value and returns its $N placeholder.
func (b *queryBuilder) arg(value any) string {
	b.args = append(b.args, value)
	return fmt.Sprintf("$%d", len(b.args))
}

func (b *queryBuilder) cond(expression string) {
	b.conditions = append(b.conditions, expression)
}

// where renders " where ..." or an empty string when no filters were applied.
func (b *queryBuilder) where() string {
	if len(b.conditions) == 0 {
		return ""
	}
	return " where " + strings.Join(b.conditions, " and ")
}

// stringFilter returns the LIKE pattern for a case-insensitive contains match.
func stringFilter(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	// Escape LIKE wildcards so user input is treated literally.
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + replacer.Replace(value) + "%", true
}

// notDeleted is the soft-delete guard. Kept as a constant so every read path
// uses the same predicate; `= null` would always be false in SQL.
const notDeleted = `deleted_at is null`

func addFilters(b *queryBuilder, f domain.Filter) {
	// Soft-deleted rows are never listed. This is unconditional so no caller can
	// forget it; the base condition is always the first term of the WHERE clause.
	b.cond(notDeleted)

	if search, ok := stringFilter(f.Search); ok {
		// One arg, reused across the three columns.
		ph := b.arg(search)
		b.cond(fmt.Sprintf("(name ilike %s or description ilike %s or slug ilike %s)", ph, ph, ph))
	}
	if slug, ok := stringFilter(f.Slug); ok {
		b.cond(fmt.Sprintf("slug ilike %s", b.arg(slug)))
	}
	if f.MinPrice != nil {
		b.cond(fmt.Sprintf("price_minor >= %s", b.arg(*f.MinPrice)))
	}
	if f.MaxPrice != nil {
		b.cond(fmt.Sprintf("price_minor <= %s", b.arg(*f.MaxPrice)))
	}
}

// orderBy whitelists the sort column so nothing from the request reaches the SQL.
func orderBy(sort domain.ProductSort) string {
	switch sort {
	case domain.SortOldest:
		return " order by created_at asc, id asc"
	case domain.SortPriceAsc:
		return " order by price_minor asc, id asc"
	case domain.SortPriceDesc:
		return " order by price_minor desc, id asc"
	default: // newest
		return " order by created_at desc, id asc"
	}
}
