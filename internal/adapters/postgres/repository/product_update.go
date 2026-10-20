package repository

import (
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/joshua-takyi/stockflow-api/internal/domain/product"
)

// buildProductUpdate renders the UPDATE statement for a partial update.
//
// Only fields set on the update are written. Column order is fixed so the SQL
// and the argument order are deterministic. Values come from product.Update's
// typed pointers, never from raw request text, so no value can be read as SQL;
// when the name changes, the slug is derived from it server-side.
//
// The returned args line up with $1..$N, with the id as the final argument.
func buildProductUpdate(id string, u product.Update) (query string, args []any) {
	sets := make([]string, 0, 5)
	args = make([]any, 0, 6)

	set := func(column string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	if u.Name != nil {
		// slug is derived from the name, so it moves with it.
		set("name", *u.Name)
		set("slug", slug.Make(*u.Name))
	}
	if u.Description != nil {
		set("description", *u.Description)
	}
	if u.PriceMinor != nil {
		set("price_minor", *u.PriceMinor)
	}
	if u.StockQuantity != nil {
		set("stock_quantity", *u.StockQuantity)
	}

	sets = append(sets, "updated_at = now()")

	// id goes last, so its placeholder is the final positional parameter.
	// The soft-delete guard means a deleted row can't be updated back into view.
	args = append(args, id)
	query = `update products set ` + strings.Join(sets, ", ") +
		fmt.Sprintf(` where id = $%d and `, len(args)) + notDeleted +
		` returning ` + productColumns
	return query, args
}
