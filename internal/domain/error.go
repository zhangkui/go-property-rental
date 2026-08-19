package domain

import "errors"

var (
	ErrNotFound  = errors.New("resource not found")
	ErrForbidden = errors.New("permission denied")
	ErrConflict  = errors.New("business conflict")
	ErrInvalid   = errors.New("invalid request")
)
