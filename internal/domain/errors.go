package domain

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrConflict           = errors.New("conflict")
	ErrDbInitialization   = errors.New("pgxpool database connection failed")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDbMigration        = errors.New("database migration error")
)
