package model

import "time"

type Book struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	PublishedYear *int      `json:"published_year"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BookInput struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	PublishedYear *int   `json:"published_year"`
}

type BookList struct {
	Books    []Book `json:"books"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}
