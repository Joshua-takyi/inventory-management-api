package repository

import (
	"strings"
	"testing"

	"github.com/joshua-takyi/stockflow-api/internal/domain"
)

func intPtr(v int) *int { return &v }

func TestAddFiltersPlaceholderNumbering(t *testing.T) {
	b := &queryBuilder{}
	addFilters(b, domain.Filter{
		Search:   "shoe",
		Slug:     "run",
		MinPrice: intPtr(100),
		MaxPrice: intPtr(5000),
	})

	// search is a single arg reused for 3 columns, then slug, min, max.
	if len(b.args) != 4 {
		t.Fatalf("expected 4 args, got %d: %v", len(b.args), b.args)
	}
	for i, arg := range b.args {
		if arg == nil {
			t.Fatalf("arg %d is nil", i)
		}
	}

	want := " where deleted_at is null and (name ilike $1 or description ilike $1 or slug ilike $1) and slug ilike $2 and price_minor >= $3 and price_minor <= $4"
	if got := b.where(); got != want {
		t.Errorf("where mismatch\n got: %s\nwant: %s", got, want)
	}
}

// With no user filters the clause must still exclude soft-deleted rows.
func TestWhereAlwaysExcludesDeletedWhenNoFilters(t *testing.T) {
	b := &queryBuilder{}
	addFilters(b, domain.Filter{})
	if got := b.where(); got != " where deleted_at is null" {
		t.Errorf("expected the soft-delete guard, got %q", got)
	}
	if len(b.args) != 0 {
		t.Errorf("expected no args, got %v", b.args)
	}
}

// A caller who passes only a search term still gets the soft-delete guard.
func TestWhereKeepsSoftDeleteGuardAlongsideFilters(t *testing.T) {
	b := &queryBuilder{}
	addFilters(b, domain.Filter{Search: "shoe"})

	where := b.where()
	if !strings.HasPrefix(where, " where deleted_at is null and ") {
		t.Errorf("soft-delete guard must lead the clause, got %q", where)
	}
}

// The count query and the page query must share the exact same filter args,
// with limit/offset appended afterwards.
func TestFilterArgsAreReusableAcrossCountAndSelect(t *testing.T) {
	b := &queryBuilder{}
	addFilters(b, domain.Filter{Search: "shoe"})

	count := b.where()
	page := domain.Pagination{Page: 3, PerPage: 20}

	// Mimic the repository: append limit/offset to a copy of the filter args.
	args := append(b.args[:len(b.args):len(b.args)], page.PerPage, page.Offset())

	countArgs := b.args[:len(b.args):len(b.args)]
	if len(countArgs) != 1 {
		t.Fatalf("count args should stay at the filter count, got %d", len(countArgs))
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args after paging, got %d", len(args))
	}
	// limit/offset placeholders continue after the filter placeholders.
	if b.where() != count {
		t.Error("where clause changed after appending paging args")
	}
	if args[1] != 20 || args[2] != 40 {
		t.Errorf("expected limit=20 offset=40, got %v %v", args[1], args[2])
	}
}

func TestStringFilterEscapesWildcards(t *testing.T) {
	got, ok := stringFilter("  50%_off\\  ")
	if !ok {
		t.Fatal("expected a filter to be produced")
	}
	want := `%50\%\_off\\%`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStringFilterRejectsBlank(t *testing.T) {
	if _, ok := stringFilter("   "); ok {
		t.Error("blank search should not produce a filter")
	}
}

func TestOrderByWhitelist(t *testing.T) {
	cases := map[domain.ProductSort]string{
		domain.SortOldest:    " order by created_at asc, id asc",
		domain.SortPriceAsc:  " order by price_minor asc, id asc",
		domain.SortPriceDesc: " order by price_minor desc, id asc",
		domain.SortNewest:    " order by created_at desc, id asc",
		"":                   " order by created_at desc, id asc",
		// Anything unrecognised (i.e. injection attempts) falls back to the default.
		"name; drop table products": " order by created_at desc, id asc",
	}
	for in, want := range cases {
		if got := orderBy(in); got != want {
			t.Errorf("orderBy(%q) = %q, want %q", in, got, want)
		}
	}
}
