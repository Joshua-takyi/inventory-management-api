package handlers

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/http/codes"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/http/dto"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/http/response"
	"github.com/joshua-takyi/stockflow-api/internal/domain"
	"github.com/joshua-takyi/stockflow-api/internal/services"
)

type ProductHandler struct {
	product *services.ProductService
	Logger  *slog.Logger
}

func NewProductHandler(p *services.ProductService, l *slog.Logger) *ProductHandler {
	return &ProductHandler{
		product: p,
		Logger:  l,
	}
}

type ResponseBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type ResponseWithData[T any] struct {
	ResponseBody
	Data T `json:"data"`
}

func (p *ProductHandler) Create() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.ProductInputDTO
		if err := ctx.ShouldBindJSON(&input); err != nil {
			response.WithData(ctx, codes.InvalidRequest, err.Error())
			return
		}

		product, err := p.product.Create(ctx, &input)

		var productErr *services.ProductError
		if err != nil {
			switch {
			case errors.As(err, &productErr):
				response.WithData(ctx, codes.InvalidRequest, productErr.Fields)
				return
			case errors.Is(err, domain.ErrAlreadyExists):
				response.WithCode(ctx, codes.Conflict)
				return
			case errors.Is(err, domain.ErrCheckViolation):
				response.WithCode(ctx, codes.Conflict)
				return
			default:
				response.Error(ctx, err)
				return
			}
		}
		response.OK(ctx, product)
	}
}

func (p *ProductHandler) List() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var query dto.ProductListQueryDTO
		if err := ctx.ShouldBindQuery(&query); err != nil {
			response.WithData(ctx, codes.InvalidRequest, err.Error())
			return
		}

		result, err := p.product.ListProducts(ctx, query)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidInput):
				response.WithCode(ctx, codes.InvalidRequest)
				return
			default:
				response.Error(ctx, err)
				return
			}
		}
		response.OK(ctx, result)
	}
}

func (p *ProductHandler) UpdateProduct() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		if id == "" {
			response.WithCode(ctx, codes.InvalidRequest)
			return
		}

		var input dto.ProductUpdateDTO
		if err := ctx.ShouldBindJSON(&input); err != nil {
			response.WithData(ctx, codes.InvalidRequest, err.Error())
			return
		}

		product, err := p.product.UpdateProduct(ctx, &input, id)
		if err != nil {
			var productErr *services.ProductError
			switch {
			case errors.As(err, &productErr):
				response.WithData(ctx, codes.InvalidRequest, productErr.Fields)
				return
			case errors.Is(err, domain.ErrNoFieldsToUpdate):
				response.WithCode(ctx, codes.InvalidRequest)
				return
			case errors.Is(err, domain.ErrInvalidInput):
				response.WithCode(ctx, codes.InvalidRequest)
				return
			case errors.Is(err, domain.ErrNotFound):
				response.WithCode(ctx, codes.NotFound)
				return
			case errors.Is(err, domain.ErrInvalidUUID):
				response.WithCode(ctx, codes.InvalidUUIDSyntax)
				return
			case errors.Is(err, domain.ErrConflict):
				response.WithCode(ctx, codes.Conflict)
				return
			case errors.Is(err, domain.ErrCheckViolation):
				response.WithCode(ctx, codes.Conflict)
				return
			default:
				response.Error(ctx, err)
				return
			}
		}
		response.OK(ctx, product)
	}
}

func (p *ProductHandler) GetByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		if id == "" {
			response.WithCode(ctx, codes.InvalidRequest)
			return
		}
		product, err := p.product.GetByID(ctx, dto.ProductGetByIdDTO{ID: id})
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				response.WithCode(ctx, codes.NotFound)
				return
			case errors.Is(err, domain.ErrInvalidUUID):
				response.WithCode(ctx, codes.InvalidUUIDSyntax)
				return
			default:
				response.Error(ctx, err)
				return
			}
		}
		response.OK(ctx, product)
	}
}

func (p *ProductHandler) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param(strings.TrimSpace("id"))
		if id == "" {
			response.WithCode(ctx, codes.InvalidRequest)
			return
		}
		_, err := p.product.Delete(ctx, id)
		if err != nil {
			// p.Logger.ErrorContext(ctx, "delete product", "error", err.Error())
			switch {
			case errors.Is(err, domain.ErrNotFound):
				response.WithCode(ctx, codes.NotFound)
				return
			default:
				response.Error(ctx, err)
				return
			}
		}
		response.OK(ctx, "product deleted")
	}
}
