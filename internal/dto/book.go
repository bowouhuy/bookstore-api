package dto

import "github.com/bowouhuy/bookstore-api/internal/entities"

type BookRequest struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	PublishedYear *int   `json:"published_year"`
}

type BookList struct {
	Books    []entities.Book
	Page     int
	PageSize int
}
