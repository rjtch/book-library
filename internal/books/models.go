package books

import (
	"time"
)

// Book represents a book in our system.
type Book struct {
	ID          string    `db:"book_id,omitempty" json:"id"`
	Title       string    `db:"title" json:"title"`
	ISBN        string    `db:"isbn" json:"isbn"`
	Category    string    `db:"category" json:"category"`
	Description string    `db:"description" json:"description"`
	Authors     string    `db:"authors" json:"authors"`
	Quantity    int       `db:"quantity" json:"quantity"`
	DateCreated time.Time `db:"date_created" json:"date_created"` // When the book was added.
	DateUpdated time.Time `db:"date_updated" json:"date_updated"` // When the book record was last modified.
}

//NewBook contains information needed to create a new Book.
type NewBook struct {
	Title       string `json:"title" json:"title"`
	ISBN        string `json:"isbn" json:"isbn"`
	Category    string `json:"category" json:"category"`
	Description string `json:"description" json:"description"`
	Authors     string `json:"authors" json:"authors"`
	Quantity    int    `json:"quantity"  validate:"gte=1"`
}

// UpdateBook defines what information may be provided to modify an
// existing Product. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateBook struct {
	Description *string    `json:"description" json:"description"`
	Authors     *string    `json:"authors" json:"authors"`
	Category    *string    `json:"category" json:"category"`
	Quantity    *int       `json:"quantity" validate:"omitempty,gte=1"`
	DateUpdated *time.Time `db:"date_updated" json:"date_updated"` // When the book record was last modified.
}
