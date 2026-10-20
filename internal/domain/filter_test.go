package domain

import "testing"

func TestPaginationOffset(t *testing.T) {
	cases := []struct {
		page, perPage, want int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 20, 40},
		{5, 10, 40},
	}
	for _, c := range cases {
		p := Pagination{Page: c.page, PerPage: c.perPage}
		if got := p.Offset(); got != c.want {
			t.Errorf("page %d perPage %d: offset = %d, want %d", c.page, c.perPage, got, c.want)
		}
	}
}

func TestPaginationNormalize(t *testing.T) {
	cases := []struct {
		name string
		in   Pagination
		want Pagination
	}{
		{"zero values", Pagination{}, Pagination{Page: 1, PerPage: 20}},
		{"negative page", Pagination{Page: -5, PerPage: 10}, Pagination{Page: 1, PerPage: 10}},
		{"per page too large", Pagination{Page: 1, PerPage: 5000}, Pagination{Page: 1, PerPage: 100}},
		{"valid values untouched", Pagination{Page: 4, PerPage: 25}, Pagination{Page: 4, PerPage: 25}},
	}
	for _, c := range cases {
		if got := c.in.Normalize(); got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestNewPaginatedTotalPages(t *testing.T) {
	cases := []struct {
		total, perPage, wantPages int
	}{
		{0, 20, 0},
		{1, 20, 1},
		{20, 20, 1},
		{21, 20, 2},
		{100, 20, 5},
		{101, 20, 6},
	}
	for _, c := range cases {
		p := Pagination{Page: 1, PerPage: c.perPage}
		got := NewPaginated([]int{}, c.total, p)
		if got.TotalPages != c.wantPages {
			t.Errorf("total %d perPage %d: totalPages = %d, want %d", c.total, c.perPage, got.TotalPages, c.wantPages)
		}
	}
}

func TestNewPaginatedNeverReturnsNilItems(t *testing.T) {
	// A nil slice must serialise as [] rather than null in JSON.
	got := NewPaginated[int](nil, 0, Pagination{Page: 1, PerPage: 20})
	if got.Items == nil {
		t.Error("expected non-nil items slice")
	}
	if len(got.Items) != 0 {
		t.Errorf("expected empty items, got %v", got.Items)
	}
}
