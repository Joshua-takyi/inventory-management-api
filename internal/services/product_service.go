package services

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gosimple/slug"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/http/dto"
	"github.com/joshua-takyi/stockflow-api/internal/domain"
	"github.com/joshua-takyi/stockflow-api/internal/domain/product"
	"github.com/joshua-takyi/stockflow-api/internal/ports"
)

var validate = validator.New()

type ProductService struct {
	product ports.ProductRepository
}

func NewProductService(p ports.ProductRepository) *ProductService {
	return &ProductService{
		product: p,
	}
}

type ProductError struct {
	Fields map[string]string
}

func (p *ProductError) Error() string {
	return "validation error"
}

// jsonFieldName maps a struct field name to its wire name via the json tag, so
// errors are keyed by what the client actually sent.
func jsonFieldName(s any, namespace string) string {
	field, ok := reflect.TypeOf(s).Elem().FieldByName(namespace)
	if !ok {
		return strings.ToLower(namespace)
	}
	name := strings.Split(field.Tag.Get("json"), ",")[0]
	if name == "" || name == "-" {
		return strings.ToLower(namespace)
	}
	return name
}

// messageFor turns a validator tag into a client-safe sentence.
func messageFor(tag, param string) string {
	switch tag {
	case "required":
		return "is required"
	case "min", "gte":
		return "must be greater than or equal to " + param
	case "max", "lte":
		return "must be less than or equal to " + param
	case "oneof":
		return "must be one of: " + strings.ReplaceAll(param, " ", ", ")
	default:
		return "is invalid"
	}
}

func validateFields(f *dto.ProductInputDTO) error {
	return collectFieldErrors(f, validate.Struct(f))
}

// validateUpdate reports field errors for the fields actually present.
func validateUpdate(u *dto.ProductUpdateDTO) error {
	return collectFieldErrors(u, validate.Struct(u))
}

// collectFieldErrors converts validator output into a field -> message map keyed
// by JSON name. validator groups multiple failures for one field into a single
// error, so each field appears once.
func collectFieldErrors(input any, err error) error {
	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return nil
	}

	fields := make(map[string]string, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		name := jsonFieldName(input, fieldErr.StructField())
		if _, exists := fields[name]; exists {
			continue
		}
		fields[name] = messageFor(fieldErr.Tag(), fieldErr.Param())
	}
	return &ProductError{Fields: fields}
}

func (p *ProductService) Create(ctx context.Context, input *dto.ProductInputDTO) (*dto.ProductResponseDTO, error) {
	if err := validateFields(input); err != nil {
		return nil, fmt.Errorf("validation of product fields %w", err)
	}

	data := &product.Product{
		Name:          input.Name,
		Slug:          generateSlug(input.Name),
		Description:   input.Description,
		PriceMinor:    input.PriceMinor,
		StockQuantity: input.StockQuantity,
	}

	created, err := p.product.Create(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("product creation %w", err)
	}
	result := toResponse(created)
	return &result, nil
}

func (p *ProductService) GetByID(ctx context.Context, input dto.ProductGetByIdDTO) (*dto.ProductResponseDTO, error) {
	if input.ID == "" {
		return nil, fmt.Errorf("id field can't be empty")
	}
	product, err := p.product.GetByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("get product %w", err)
	}
	result := toResponse(product)
	return &result, nil
}

func (p *ProductService) ListProducts(ctx context.Context, input dto.ProductListQueryDTO) (*dto.ProductListResponseDTO, error) {
	if err := validate.Struct(input); err != nil {
		return nil, fmt.Errorf("validation of list query %w", domain.ErrInvalidInput)
	}
	if input.MinPrice != nil && input.MaxPrice != nil && *input.MinPrice > *input.MaxPrice {
		return nil, fmt.Errorf("min_price cannot exceed max_price: %w", domain.ErrInvalidInput)
	}

	page := domain.Pagination{Page: input.Page, PerPage: input.PerPage}.Normalize()

	filter := domain.Filter{
		Search:   input.Search,
		Slug:     input.Slug,
		MinPrice: input.MinPrice,
		MaxPrice: input.MaxPrice,
		Sort:     domain.ProductSort(input.Sort),
	}

	products, total, err := p.product.ListProducts(ctx, filter, page)
	if err != nil {
		return nil, fmt.Errorf("list products %w", err)
	}

	items := make([]dto.ProductResponseDTO, 0, len(products))
	for _, pr := range products {
		items = append(items, toResponse(pr))
	}

	result := domain.NewPaginated(items, total, page)
	return &result, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, input *dto.ProductUpdateDTO, id string) (*dto.ProductResponseDTO, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id field can't be empty: %w", domain.ErrInvalidInput)
	}

	update := input.ToUpdate()
	if update.IsEmpty() {
		return nil, fmt.Errorf("update needs at least one field: %w", domain.ErrNoFieldsToUpdate)
	}

	if err := validateUpdate(input); err != nil {
		return nil, fmt.Errorf("validation of product fields %w", err)
	}

	updated, err := p.product.UpdateProduct(ctx, id, update)
	if err != nil {
		return nil, fmt.Errorf("update product %w", err)
	}

	result := toResponse(updated)
	return &result, nil
}

func (p *ProductService) Delete(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("id field can't be empty")
	}
	return p.product.Delete(ctx, id)
}

func toResponse(p *product.Product) dto.ProductResponseDTO {
	return dto.ProductResponseDTO{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Slug:          p.Slug,
		PriceMinor:    p.PriceMinor,
		StockQuantity: p.StockQuantity,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func generateSlug(name string) string {
	return slug.Make(name)
}
