package response

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/http/codes"
	"github.com/joshua-takyi/stockflow-api/internal/domain"
)

type ErrorBody struct {
	Code    codes.Code `json:"code"`
	Message string     `json:"message"`
}

// resolve maps domain errors to a response code. Order matters: most specific first.
func resolve(err error) codes.Code {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return codes.NotFound
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAlreadyExists):
		return codes.Conflict
	case errors.Is(err, domain.ErrInvalidCredentials):
		return codes.InvalidCredentials
	case errors.Is(err, domain.ErrUnauthorized):
		return codes.Unauthorized
	case errors.Is(err, domain.ErrInvalidInput):
		return codes.InvalidRequest
	case errors.Is(err, context.DeadlineExceeded):
		return codes.Timeout
	default:
		return codes.InternalError
	}
}

// Error logs the full wrapped chain once, then returns a safe body to the client.
func Error(c *gin.Context, err error) {
	code := resolve(err)

	if code == codes.InternalError {
		slog.Error("request failed", "err", err, "path", c.FullPath())
	}

	c.AbortWithStatusJSON(code.Status(), ErrorBody{
		Code:    code,
		Message: code.Message(),
	})
}

// WithCode is for errors that don't come from the domain (bad JSON, missing header, etc.).
func WithCode(c *gin.Context, code codes.Code) {
	c.AbortWithStatusJSON(code.Status(), ErrorBody{
		Code:    code,
		Message: code.Message(),
	})
}

func OK(c *gin.Context, data any) { c.JSON(http.StatusOK, data) }
