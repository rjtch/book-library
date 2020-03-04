package category

import (
	"time"
)

//BookCategory represents the category in which a book has to be ranged
type BookCategory struct {
	ID               string    `db:"category_id,omitempty" json:"id"`
	CategoryName     string    `db:"name,omitempty" json:"name"`
	NumberOfBooksIn  int       `db:"books_in" json:"books_in"`
	NumberOfBooksOut int       `db:"books_out" json:"books_out"`
	DateCreated      time.Time `db:"date_created" json:"date_created"` // When the bookCategory was added.
	DateUpdated      time.Time `db:"date_updated" json:"date_updated"` // When the bookCategory record was last modified.
}

//NewBookCategory represents a new created category in which a book has to be ranged
type NewBookCategory struct {
	CategoryName     string    `json:"name,omitempty" json:"name"`
	NumberOfBooksIn  int       `json:"books_in" json:"books_in"`
	NumberOfBooksOut int       `json:"books_out" json:"books_out"`
	DateCreated      time.Time `json:"date_created" json:"date_created"`
}

//UpdateBookCategory updates an existing category
type UpdateBookCategory struct {
	CategoryName     *string    `json:"name"`
	NumberOfBooksIn  *int       `json:"books_in"`
	NumberOfBooksOut *int       `json:"books_out"`
	DateUpdated      *time.Time `json:"date_updated" json:"date_updated"` // When the bookCategory record was last modified.
}
