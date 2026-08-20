package model

import (
	"errors"
	"fmt"
)

// Sentinel errors. Wrapped by store/service; surfaced as HTTP status via HTTPCode().
var (
	ErrNotFound      = errors.New("not found")
	ErrInvalid       = errors.New("invalid input")
	ErrStateConflict = errors.New("state conflict")
	ErrInvariant     = errors.New("invariant violation")
	ErrDuplicate     = errors.New("duplicate")
)

// HTTPCode maps a sentinel error to an HTTP status code.
func HTTPCode(err error) int {
	switch {
	case err == nil:
		return 200
	case errors.Is(err, ErrNotFound):
		return 404
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrInvariant):
		return 422
	case errors.Is(err, ErrStateConflict), errors.Is(err, ErrDuplicate):
		return 409
	default:
		return 500
	}
}

// wrap returns err with a message prefix, preserving sentinel identity for Is().
func wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}
