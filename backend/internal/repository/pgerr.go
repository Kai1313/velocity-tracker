package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"velocity-tracker/backend/internal/apperr"
)

// wrapReadErr maps a failed lookup (Get/Update/Delete targeting a missing
// row) onto apperr.ErrNotFound.
func wrapReadErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.ErrNotFound
	}
	return err
}

// wrapMutationErr maps constraint violations from an INSERT/UPDATE/DELETE
// onto the sentinel the caller expects. A foreign-key violation means two
// different things depending on direction: on insert/update it means the
// caller pointed at something that doesn't exist (a validation problem);
// on delete it means some other row still points at this one (a conflict).
func wrapMutationErr(err error, isDelete bool) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			if isDelete {
				return fmt.Errorf("%w: still referenced by another record (%s)", apperr.ErrConflict, pgErr.ConstraintName)
			}
			return fmt.Errorf("%w: references a record that does not exist (%s)", apperr.ErrValidation, pgErr.ConstraintName)
		case pgerrcode.UniqueViolation:
			return fmt.Errorf("%w: already exists (%s)", apperr.ErrConflict, pgErr.ConstraintName)
		case pgerrcode.CheckViolation, pgerrcode.NotNullViolation:
			return fmt.Errorf("%w: %s", apperr.ErrValidation, pgErr.Message)
		case pgerrcode.RaiseException:
			// Custom domain triggers (e.g. "one open SprintEntry per ticket")
			// signal their invariant violations via RAISE EXCEPTION.
			return fmt.Errorf("%w: %s", apperr.ErrConflict, pgErr.Message)
		}
	}
	return wrapReadErr(err)
}
