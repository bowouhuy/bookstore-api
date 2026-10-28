package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/bowouhuy/bookstore-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const bookColumns = `id, title, author, published_year, created_at, updated_at`

type BookRepository struct {
	db *pgxpool.Pool
}

func NewBookRepository(db *pgxpool.Pool) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) Create(ctx context.Context, input model.BookInput) (*model.Book, error) {
	query := `
		INSERT INTO books (title, author, published_year)
		VALUES ($1, $2, $3)
		RETURNING ` + bookColumns

	row := r.db.QueryRow(ctx, query, input.Title, input.Author, input.PublishedYear)

	book, err := scanBook(row)
	if err != nil {
		return nil, fmt.Errorf("insert book: %w", err)
	}
	return book, nil
}

func (r *BookRepository) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	query := `SELECT ` + bookColumns + ` FROM books WHERE id = $1`

	book, err := scanBook(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get book by id: %w", err)
	}
	return book, nil
}

func (r *BookRepository) List(ctx context.Context, limit, offset int) ([]model.Book, error) {
	query := `SELECT ` + bookColumns + ` FROM books ORDER BY id LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	books := make([]model.Book, 0)
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

func (r *BookRepository) Update(ctx context.Context, id int64, input model.BookInput) (*model.Book, error) {
	query := `
		UPDATE books
		SET title = $1, author = $2, published_year = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING ` + bookColumns

	row := r.db.QueryRow(ctx, query, input.Title, input.Author, input.PublishedYear, id)

	book, err := scanBook(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("update book: %w", err)
	}
	return book, nil
}

func (r *BookRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM books WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete book: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func scanBook(row rowScanner) (*model.Book, error) {
	var b model.Book
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
