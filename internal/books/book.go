package books

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	auth "github.com/book-library/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const booksCollection = "books"

//An instance of the book-category repository to update book's quantity in the database
//var ctg *category.BookCategoryRepository

type key int

var categoryKey key

// Predefined errors identify expected failure conditions.
var (
	// ErrNotFound is used when a specific Product is requested but does not exist.
	ErrNotFound = errors.New("Book not found")

	// ErrInvalidID is used when an invalid UUID is provided.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a users tries to do something that is forbidden to
	// them according to our access control policies.
	ErrForbidden = errors.New("Attempted action is not allowed")
)

//List retrieves a list of existing books from the database
func List(ctx context.Context, db *sqlx.DB) ([]Book, error) {
	ctx, span := trace.StartSpan(ctx, "internal.book.List")
	defer span.End()

	books := []Book{}
	const q = `SELECT * FROM books`

	if err := db.SelectContext(ctx, &books, q); err != nil {
		return nil, errors.Wrap(err, "selecting books")
	}

	return books, nil
}

//Retrieve gets the specific book from the database
func Retrieve(ctx context.Context, id string, db *sqlx.DB) (*Book, error) {
	ctx, span := trace.StartSpan(ctx, "internal.book.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var b Book
	const q = `SELECT * FROM books WHERE book_id = $1`
	if err := db.GetContext(ctx, &b, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting category %q", id)
	}

	return &b, nil
}

//Retrieve gets the specific book from the database
func RetrieveByTitle(ctx context.Context, title string, db *sqlx.DB) (*Book, error) {
	ctx, span := trace.StartSpan(ctx, "internal.book.RetrieveByTitle")
	defer span.End()

	var b Book
	const q = `SELECT * FROM books WHERE title = $1`
	if err := db.GetContext(ctx, &b, q, title); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting category %q", title)
	}

	return &b, nil
}

// Create inserts a new book into the database.
func Create(ctx context.Context, now time.Time, n NewBook, user auth.Claims, db *sqlx.DB) (*Book, error) {
	ctx, span := trace.StartSpan(ctx, "internal.book.Create")
	defer span.End()

	// If you do not have the admin role ...
	// then get outta here!
	if !user.HasRole(auth.RoleAdmin) {
		return nil, ErrForbidden
	}

	book := Book{
		ID:          uuid.New().String(),
		Title:       n.Title,
		ISBN:        n.ISBN,
		Category:    n.Category,
		Description: n.Description,
		Quantity:    n.Quantity,
		Authors:     n.Authors,
		DateCreated: now.UTC(),
		DateUpdated: now.UTC(),
	}

	const q = `INSERT INTO books
		(book_id, title, isbn, category, authors, description, quantity, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := db.ExecContext(
		ctx, q,
		book.ID, book.Title, book.ISBN, book.Category,book.Authors, book.Description, book.Quantity,
		book.DateCreated, book.DateUpdated,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting book")
	}

	//catgory, errr  := category.RetrieveByCategory(ctx, db, book.Category)
	//if (errr != nil) {
	//	return nil, errors.Wrap(errr, "category might not exist ")
	//} else {
	//	catgory.NumberOfBooksIn++
	//}

	rbk, errR := Retrieve(ctx, book.ID, db)

	if errR != nil {
		fmt.Printf(":...rb %v", rbk)
	}

	rbks, errRs := List(ctx, db)

	if errRs != nil {
		fmt.Printf(":...rb %v", rbks)
	}

	return &book, nil
}

// Update replaces a book document in the database.
func Update(ctx context.Context, id string, upd UpdateBook, now time.Time, user auth.Claims, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.book.Update")
	defer span.End()

	// If you do not have the admin role ...
	// then get outta here!
	if !user.HasRole(auth.RoleAdmin) {
		return ErrForbidden
	}

	book, err := Retrieve(ctx, id, db)

	if err != nil {
		return err
	}

	if upd.Quantity != nil {
		book.Quantity = *upd.Quantity
	}

	if upd.Description != nil {
		book.Description = *upd.Description
	}

	if upd.Authors != nil {
		book.Authors = *upd.Authors
	}

	if upd.DateUpdated != nil {
		book.DateUpdated = now
	}

	const q = `UPDATE books SET
	"authors" = $2,
	"description" = $3,
	"quantity" = $4
	WHERE book_id = $1`
	_, err = db.ExecContext(ctx, q, id,
		book.Authors, book.Description, book.Quantity,
	)

	bks2, errB2 := List(ctx, db)
	if errB2 != nil {
		fmt.Printf("%v", bks2)
	}
	if err != nil {
		return errors.Wrap(err, "updating book")
	}

	return nil
}

// Delete removes a book from the database.
func Delete(ctx context.Context, id string, user auth.Claims, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.book.Delete")
	defer span.End()

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	if !user.HasRole(auth.RoleAdmin) && user.Subject != id {
		return ErrForbidden
	}

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM books WHERE book_id = $1`

	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting book %s", id)
	}

	return nil
}
