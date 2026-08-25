// Package apperr defines sentinel errors that carry HTTP-status intent
// across the repository -> service -> handler boundary without those
// layers importing net/http.
package apperr

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrValidation = errors.New("validation failed")
)
