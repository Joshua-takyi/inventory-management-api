package repository

import (
	"strings"
	"testing"

	"github.com/joshua-takyi/stockflow-api/internal/domain/product"
)

func strPtr(v string) *string { return &v }

func TestBuildProductUpdateFull(t *testing.T) {
	query, args := buildProductUpdate("abc-id", product.Update{
		Name:          strPtr("New Name"),
		Description:   strPtr("desc"),
		PriceMinor:    intPtr(500),
		StockQuantity: intPtr(7),
	})

	want := "update products set name = $1, slug = $2, description = $3, price_minor = $4, stock_quantity = $5, updated_at = now() where id = $6 and deleted_at is null returning " + productColumns
	if query != want {
		t.Errorf("query mismatch\n got: %s\nwant: %s", query, want)
	}

	if len(args) != 6 {
		t.Fatalf("expected 6 args, got %d: %v", len(args), args)
	}
	// Values must line up with $1..$6, and id must be last.
	if args[0] != "New Name" {
		t.Errorf("$1 = %v, want the name", args[0])
	}
	if args[1] != "new-name" {
		t.Errorf("$2 = %v, want the derived slug", args[1])
	}
	if args[2] != "desc" || args[3] != 500 || args[4] != 7 {
		t.Errorf("payload args wrong: %v", args[2:5])
	}
	if args[5] != "abc-id" {
		t.Errorf("$6 = %v, want the id", args[5])
	}
}

// The id must track the SET placeholder count, not a fixed index. A single-field
// update previously risked binding the id as $5 against a one-column SET clause.
func TestBuildProductUpdateIdIsAlwaysLastPlaceholder(t *testing.T) {
	cases := []struct {
		name      string
		update    product.Update
		wantQuery string
		wantArgs  int
	}{
		{
			name:      "stock quantity only",
			update:    product.Update{StockQuantity: intPtr(3)},
			wantQuery: "update products set stock_quantity = $1, updated_at = now() where id = $2 and deleted_at is null returning " + productColumns,
			wantArgs:  2,
		},
		{
			name:      "price only",
			update:    product.Update{PriceMinor: intPtr(99)},
			wantQuery: "update products set price_minor = $1, updated_at = now() where id = $2 and deleted_at is null returning " + productColumns,
			wantArgs:  2,
		},
		{
			name:      "description only",
			update:    product.Update{Description: strPtr("d")},
			wantQuery: "update products set description = $1, updated_at = now() where id = $2 and deleted_at is null returning " + productColumns,
			wantArgs:  2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			query, args := buildProductUpdate("the-id", c.update)
			if query != c.wantQuery {
				t.Errorf("query mismatch\n got: %s\nwant: %s", query, c.wantQuery)
			}
			if len(args) != c.wantArgs {
				t.Fatalf("expected %d args, got %d: %v", c.wantArgs, len(args), args)
			}
			if args[len(args)-1] != "the-id" {
				t.Errorf("last arg = %v, want the id", args[len(args)-1])
			}
		})
	}
}

func TestBuildProductUpdateNameOnlyDerivesSlug(t *testing.T) {
	query, args := buildProductUpdate("id", product.Update{Name: strPtr("Fresh Sneaker")})

	want := "update products set name = $1, slug = $2, updated_at = now() where id = $3 and deleted_at is null returning " + productColumns
	if query != want {
		t.Errorf("query mismatch\n got: %s\nwant: %s", query, want)
	}
	if args[1] != "fresh-sneaker" {
		t.Errorf("slug arg = %v, want fresh-sneaker", args[1])
	}
}

// Every emitted placeholder must have exactly one argument behind it.
// A deleted product must not be updatable back into view.
func TestBuildProductUpdateAlwaysGuardsSoftDeleted(t *testing.T) {
	query, _ := buildProductUpdate("id", product.Update{Name: strPtr("n")})
	if !strings.Contains(query, "and deleted_at is null") {
		t.Errorf("update must exclude soft-deleted rows, got %s", query)
	}
}

func TestBuildProductUpdatePlaceholderCountMatchesArgs(t *testing.T) {
	query, args := buildProductUpdate("id", product.Update{
		Name:          strPtr("n"),
		PriceMinor:    intPtr(1),
		StockQuantity: intPtr(2),
	})

	// Highest $N in the query must equal the number of args.
	highest := 0
	for _, tok := range strings.FieldsFunc(query, func(r rune) bool { return r == '$' || r == ' ' || r == ',' }) {
		if n := atoi(tok); n > highest {
			highest = n
		}
	}
	if highest != len(args) {
		t.Errorf("highest placeholder $%d but %d args: %v", highest, len(args), args)
	}
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}
