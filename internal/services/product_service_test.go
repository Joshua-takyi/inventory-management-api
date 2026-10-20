package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/joshua-takyi/stockflow-api/internal/adapters/http/dto"
	"github.com/joshua-takyi/stockflow-api/internal/domain"
	"github.com/joshua-takyi/stockflow-api/internal/domain/product"
)

// stubRepo records the update it was handed so the service mapping can be asserted.
type stubRepo struct {
	gotID     string
	gotUpdate product.Update
	calls     int
	returned  *product.Product
	err       error
}

func (s *stubRepo) Create(context.Context, *product.Product) (*product.Product, error) {
	return nil, errors.New("not used")
}

func (s *stubRepo) GetByID(context.Context, string) (*product.Product, error) {
	return nil, errors.New("not used")
}

func (s *stubRepo) ListProducts(context.Context, domain.Filter, domain.Pagination) ([]*product.Product, int, error) {
	return nil, 0, errors.New("not used")
}

func (s *stubRepo) Delete(ctx context.Context, id string) (string, error) {
	return "", errors.New("not used")
}

func (s *stubRepo) UpdateProduct(_ context.Context, id string, u product.Update) (*product.Product, error) {
	s.calls++
	s.gotID = id
	s.gotUpdate = u
	if s.err != nil {
		return nil, s.err
	}
	if s.returned != nil {
		return s.returned, nil
	}
	return &product.Product{ID: [16]byte{}}, nil
}

func strPtr(v string) *string { return &v }
func intPtr(v int) *int       { return &v }

func TestUpdateProductRejectsEmptyID(t *testing.T) {
	repo := &stubRepo{}
	svc := NewProductService(repo)

	_, err := svc.UpdateProduct(context.Background(), &dto.ProductUpdateDTO{Name: strPtr("x")}, "   ")

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if repo.calls != 0 {
		t.Error("repository must not be called when the id is blank")
	}
}

func TestUpdateProductRejectsEmptyBody(t *testing.T) {
	repo := &stubRepo{}
	svc := NewProductService(repo)

	_, err := svc.UpdateProduct(context.Background(), &dto.ProductUpdateDTO{}, "some-id")

	if !errors.Is(err, domain.ErrNoFieldsToUpdate) {
		t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
	}
	if repo.calls != 0 {
		t.Error("repository must not be called for an empty update")
	}
}

func TestCreateValidationMessageIsClientSafe(t *testing.T) {
	repo := &stubRepo{}
	svc := NewProductService(repo)

	// Exactly the payload that produced a garbled message: negative price.
	input := &dto.ProductInputDTO{
		Name:          "nemo",
		Description:   "lorem1 lorem",
		PriceMinor:    -400,
		StockQuantity: 100,
	}

	_, err := svc.Create(context.Background(), input)

	var productErr *ProductError
	if !errors.As(err, &productErr) {
		t.Fatalf("expected *ProductError, got %v", err)
	}

	// Keyed by JSON name, not by validator's sentence.
	if _, ok := productErr.Fields["price_minor"]; !ok {
		t.Fatalf("expected a price_minor error, got %v", productErr.Fields)
	}
	if got := productErr.Fields["price_minor"]; got != "must be greater than or equal to 0" {
		t.Errorf("message = %q, want %q", got, "must be greater than or equal to 0")
	}
	// validator would have reported the field as an int bound to the trip metric,
	// so the raw tag must not leak into the key or the message.
	for key, msg := range productErr.Fields {
		if strings.Contains(key, "Key:") || strings.Contains(msg, "failed on the") {
			t.Errorf("raw validator output leaked: %q: %q", key, msg)
		}
	}
}

// Multiple bad fields should each get their own entry, keyed by JSON name.
func TestCreateValidationReportsEachField(t *testing.T) {
	repo := &stubRepo{}
	svc := NewProductService(repo)

	_, err := svc.Create(context.Background(), &dto.ProductInputDTO{
		Name:          "",
		Description:   "",
		PriceMinor:    -1,
		StockQuantity: -2,
	})

	var productErr *ProductError
	if !errors.As(err, &productErr) {
		t.Fatalf("expected *ProductError, got %v", err)
	}

	expect := []string{"name", "description", "price_minor", "stock_quantity"}
	for _, field := range expect {
		if _, ok := productErr.Fields[field]; !ok {
			t.Errorf("missing error for %q, got %v", field, productErr.Fields)
		}
	}
}

func TestUpdateProductRejectsValidationErrors(t *testing.T) {
	cases := []struct {
		name  string
		input dto.ProductUpdateDTO
	}{
		{"negative price", dto.ProductUpdateDTO{PriceMinor: intPtr(-1)}},
		{"negative stock", dto.ProductUpdateDTO{StockQuantity: intPtr(-5)}},
		{"blank name", dto.ProductUpdateDTO{Name: strPtr("")}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := &stubRepo{}
			svc := NewProductService(repo)

			_, err := svc.UpdateProduct(context.Background(), &c.input, "some-id")

			var productErr *ProductError
			if !errors.As(err, &productErr) {
				t.Fatalf("expected *ProductError, got %v", err)
			}
			if len(productErr.Fields) == 0 {
				t.Error("expected at least one field error")
			}
			if repo.calls != 0 {
				t.Error("repository must not be called when validation fails")
			}
		})
	}
}

// A PATCH naming only price_minor must leave every other domain field nil so the
// repository writes just that column.
func TestUpdateProductMapsOnlyProvidedFields(t *testing.T) {
	repo := &stubRepo{}
	svc := NewProductService(repo)

	input := &dto.ProductUpdateDTO{PriceMinor: intPtr(1234)}
	if _, err := svc.UpdateProduct(context.Background(), input, "prod-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.gotID != "prod-1" {
		t.Errorf("id = %q, want prod-1", repo.gotID)
	}
	u := repo.gotUpdate
	if u.PriceMinor == nil || *u.PriceMinor != 1234 {
		t.Errorf("price_minor not mapped: %v", u.PriceMinor)
	}
	if u.Name != nil || u.Description != nil || u.StockQuantity != nil {
		t.Errorf("omitted fields must stay nil, got %+v", u)
	}
}

func TestUpdateProductPropagatesRepositoryError(t *testing.T) {
	repo := &stubRepo{err: domain.ErrNotFound}
	svc := NewProductService(repo)

	_, err := svc.UpdateProduct(context.Background(), &dto.ProductUpdateDTO{Name: strPtr("n")}, "missing")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound to survive wrapping, got %v", err)
	}
}
