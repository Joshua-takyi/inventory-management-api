package codes

import "net/http"

type Code string

const (
	InternalError      Code = "INTERNAL_ERROR"
	InvalidRequest     Code = "INVALID_REQUEST"
	ValidationFailed   Code = "VALIDATION_FAILED"
	Unauthorized       Code = "UNAUTHORIZED"
	InvalidCredentials Code = "INVALID_CREDENTIALS"
	Forbidden          Code = "FORBIDDEN"
	NotFound           Code = "NOT_FOUND"
	Conflict           Code = "CONFLICT"
	TooManyRequests    Code = "TOO_MANY_REQUESTS"
	Timeout            Code = "REQUEST_TIMEOUT"
)

type info struct {
	status  int
	message string
}

var registry = map[Code]info{
	InternalError:      {http.StatusInternalServerError, "internal server error, try again later"},
	InvalidRequest:     {http.StatusBadRequest, "invalid request"},
	ValidationFailed:   {http.StatusUnprocessableEntity, "validation failed"},
	Unauthorized:       {http.StatusUnauthorized, "authentication required"},
	InvalidCredentials: {http.StatusUnauthorized, "invalid credentials"},
	Forbidden:          {http.StatusForbidden, "you do not have permission to perform this action"},
	NotFound:           {http.StatusNotFound, "resource not found"},
	Conflict:           {http.StatusConflict, "resource conflict"},
	TooManyRequests:    {http.StatusTooManyRequests, "too many requests, slow down"},
	Timeout:            {http.StatusGatewayTimeout, "request timed out"},
}

// Status returns the HTTP status for the code, falling back to 500.
func (c Code) Status() int {
	if i, ok := registry[c]; ok {
		return i.status
	}
	return http.StatusInternalServerError
}

// Message returns the default client-safe message, falling back to the internal one.
func (c Code) Message() string {
	if i, ok := registry[c]; ok {
		return i.message
	}
	return registry[InternalError].message
}

func (c Code) String() string { return string(c) }
