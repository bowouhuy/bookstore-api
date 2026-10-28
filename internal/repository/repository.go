package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type rowScanner interface {
	Scan(dest ...any) error
}

const pgUniqueViolation = "23505"

func IsUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == constraint
}
