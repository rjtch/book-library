package loans

import (
	"time"
)

//Loan represents a book in our system.
type Loan struct {
	ID           string    `db:"loan_id,omitempty" json:"id"`
	BookTitle    string    `db:"title" json:"title"`
	BookISBN     string    `db:"isbn" json:"isbn"`
	BookQuantity int       `db:"quantity"  json:"category"`
	BookID       string    `db:"book_id,omitempty" json:"book_id"`
	LoanDate     time.Time `db:"loan_date" json:"loan_date"` // When the Loan was added.
	ReturnDate  time.Time  `db:"date_return" json:"date_return"` // When the Loan record was last modified.
	UserID       string    `db:"user_id" json:"user_id"`
}

//NewLoan contains information needed to create a new Book.
type NewLoan struct {
	BookTitle    string `json:"title" json:"title"`
	BookISBN     string `json:"isbn" json:"isbn"`
	BookID       string `json:"book_id,omitempty" json:"book_id"`
	BookQuantity int    `json:"quantity"  validate:"gte=1"`
}

// UpdateLoan defines what information may be provided to modify an
// existing Product. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateLoan struct {
	BookISBN     *string    `json:"isbn" json:"isbn"`
	BookQuantity *int       `json:"quantity"  validate:"gte=1"`
	ReturnDate  *time.Time `json:"date_return" json:"date_return"` // When the Loan record was last modified.
}
