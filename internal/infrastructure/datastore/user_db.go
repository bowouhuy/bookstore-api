package datastore

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/bowouhuy/bookstore-api/internal/entities"
	"github.com/bowouhuy/bookstore-api/internal/entities/repository"
)

const userColumns = `id, name, email, password_hash, created_at, updated_at`

const (
	pgUniqueViolation = "23505"
	// Name Postgres auto-generates for the UNIQUE constraint on users.email.
	usersEmailConstraint = "users_email_key"
)

type userRepo struct {
	*BaseRepo
}

// NewUserRepo returns a user repository.
func NewUserRepo(base *BaseRepo) repository.UserRepo {
	return &userRepo{
		BaseRepo: base,
	}
}

func (r *userRepo) Create(ctx context.Context, name, email, passwordHash string) (*entities.User, error) {
	query := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING ` + userColumns

	user, err := scanUser(r.dbMaster.QueryRow(ctx, query, name, email, passwordHash))
	if err != nil {
		if isUniqueViolation(err, usersEmailConstraint) {
			return nil, entities.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

// GetByEmail reads from master so a login right after registering
// is not affected by replication lag.
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE email = $1`

	user, err := scanUser(r.dbMaster.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE id = $1`

	user, err := scanUser(r.dbSlave.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func scanUser(row pgx.Row) (*entities.User, error) {
	var u entities.User
	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == constraint
}
