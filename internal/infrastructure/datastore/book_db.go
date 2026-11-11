package datastore

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
	"github.com/bowouhuy/bookstore-api/internal/entities/repository"
)

const bookColumns = `id, title, author, published_year, created_at, updated_at`

type bookRepo struct {
	*BaseRepo
}

// NewBookRepo returns a book repository.
func NewBookRepo(base *BaseRepo) repository.BookRepo {
	return &bookRepo{
		BaseRepo: base,
	}
}

func (r *bookRepo) Create(ctx context.Context, input dto.BookRequest) (*entities.Book, error) {
	query := `
		INSERT INTO books (title, author, published_year)
		VALUES ($1, $2, $3)
		RETURNING ` + bookColumns

	book, err := scanBook(r.dbMaster.QueryRow(ctx, query, input.Title, input.Author, input.PublishedYear))
	if err != nil {
		return nil, fmt.Errorf("insert book: %w", err)
	}

	return book, nil
}

func (r *bookRepo) GetByID(ctx context.Context, id int64) (*entities.Book, error) {
	query := `SELECT ` + bookColumns + ` FROM books WHERE id = $1`

	book, err := scanBook(r.dbSlave.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, fmt.Errorf("get book by id: %w", err)
	}

	return book, nil
}

func (r *bookRepo) List(ctx context.Context, limit, offset int) ([]entities.Book, error) {
	query := `SELECT ` + bookColumns + ` FROM books ORDER BY id LIMIT $1 OFFSET $2`

	rows, err := r.dbSlave.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	books := make([]entities.Book, 0)
	for rows.Next() {
		book, err := scanBook(rows)
		if err != nil {
			return nil, fmt.Errorf("scan book: %w", err)
		}
		books = append(books, *book)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate books: %w", err)
	}

	return books, nil
}

func (r *bookRepo) Update(ctx context.Context, id int64, input dto.BookRequest) (*entities.Book, error) {
	query := `
		UPDATE books
		SET title = $1, author = $2, published_year = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING ` + bookColumns

	book, err := scanBook(r.dbMaster.QueryRow(ctx, query, input.Title, input.Author, input.PublishedYear, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, fmt.Errorf("update book: %w", err)
	}

	return book, nil
}

func (r *bookRepo) Delete(ctx context.Context, id int64) error {
	tag, err := r.dbMaster.Exec(ctx, `DELETE FROM books WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete book: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}

	return nil
}

func scanBook(row pgx.Row) (*entities.Book, error) {
	var b entities.Book
	err := row.Scan(
		&b.ID,
		&b.Title,
		&b.Author,
		&b.PublishedYear,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &b, nil
}
