package category

import (
	"context"
	"database/sql"
	"time"

	"github.com/book-library/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const (
	categoriesCollection = "categories"
)

type key int

var categoryKey key

// Predefined errors identify expected failure conditions.
var (
	// ErrNotFound is used when a specific Product is requested but does not exist.
	ErrNotFound = errors.New("category not found")

	// ErrInvalidID is used when an invalid UUID is provided.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a users tries to do something that is forbidden to
	// them according to our access control policies.
	ErrForbidden = errors.New("Attempted action is not allowed")
)

//List retrieves a list of existing bookCategory from the databse
func List(ctx context.Context, db *sqlx.DB) ([]BookCategory, error) {
	ctx, span := trace.StartSpan(ctx, "internal.book-category.List")
	defer span.End()

	category := []BookCategory{}
	const q = `SELECT * FROM categories`

	if err := db.SelectContext(ctx, &category, q); err != nil {
		return nil, errors.Wrap(err, "selecting category")
	}

	return category, nil
}

//Retrieve gets the specific bookCategory from the database
func Retrieve(ctx context.Context, user auth.Claims, db *sqlx.DB, id string) (*BookCategory, error) {
	ctx, span := trace.StartSpan(ctx, "internal.book-category.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	// If you do not have the admin role ...
	// then get outta here!
	if !user.HasRole(auth.RoleUser) {
		return nil, ErrForbidden
	}

	var b BookCategory
	const q = `SELECT * FROM categories WHERE category_id = $1`
	if err := db.GetContext(ctx, &b, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting category %q", id)
	}

	return &b, nil
}

//RetrieveByCategory gets the specific bookCategory from the database by categoryName
func RetrieveByCategory(ctx context.Context, user auth.Claims, db *sqlx.DB, categoryName string) (*BookCategory, error) {
	ctx, span := trace.StartSpan(ctx, "internal.book-category.Retrieve")
	defer span.End()

	// If you do not have the admin role ...
	// then get outta here!
	if !user.HasRole(auth.RoleUser) {
		return nil, ErrForbidden
	}

	var b BookCategory
	const q = `SELECT * FROM categories WHERE name = $1`
	if err := db.GetContext(ctx, &b, q, categoryName); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting category %q", categoryName)
	}

	return &b, nil
}

// Create inserts a new book into the database.
func Create(ctx context.Context, now time.Time, n NewBookCategory, user auth.Claims, db *sqlx.DB) (*BookCategory, error) {
	ctx, span := trace.StartSpan(ctx, "internal.book-category.Create")
	defer span.End()

	// If you do not have the admin role ...
	// then get outta here!
	if !user.HasRole(auth.RoleAdmin) {
		return nil, ErrForbidden
	}

	category := BookCategory{
		ID:               uuid.New().String(),
		CategoryName:     n.CategoryName,
		NumberOfBooksIn:  n.NumberOfBooksIn,
		NumberOfBooksOut: n.NumberOfBooksOut,
		DateCreated:      now.UTC(),
		DateUpdated:      now.UTC(),
	}

	const q = `INSERT INTO categories
		(category_id, name, books_in, books_out, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.ExecContext(
		ctx, q,
		category.ID, category.CategoryName, category.NumberOfBooksIn, category.NumberOfBooksOut,
		category.DateCreated, category.DateUpdated,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting category")
	}

	return &category, nil
}

// Update replaces a book-category document in the database.
func Update(ctx context.Context, id string, upd UpdateBookCategory, now time.Time, user auth.Claims, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.book-category.Update")
	defer span.End()

	// // If you do not have the admin role ...
	// // then get outta here!
	if !user.HasRole(auth.RoleAdmin) {
		return ErrForbidden
	}

	category, err := Retrieve(ctx, user, db, id)
	if err != nil {
		return err
	}

	if upd.NumberOfBooksIn != nil {
		category.NumberOfBooksIn += *upd.NumberOfBooksIn
	}

	if upd.NumberOfBooksOut != nil {
		category.NumberOfBooksOut = *upd.NumberOfBooksOut
	}

	if upd.DateUpdated != nil {
		category.DateUpdated = *upd.DateUpdated
	}

	if upd.DateUpdated != nil {
		category.DateUpdated = now
	}

	if upd.CategoryName != nil {
		category.CategoryName = *upd.CategoryName
	}

	const q = `UPDATE categories SET
		"name" = $2,
		"books_in" = $3,
		"books_out" = $4
		WHERE category_id = $1`
	_, err = db.ExecContext(ctx, q, id,
		upd.CategoryName, upd.NumberOfBooksIn, upd.NumberOfBooksOut,
	)
	if err != nil {
		return errors.Wrap(err, "updating category")
	}

	return nil
}

// Delete removes a book-category from the database.
func Delete(ctx context.Context, id string, user auth.Claims, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.book-category.Delete")
	defer span.End()

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	if !user.HasRole(auth.RoleAdmin) && user.Subject != id {
		return ErrForbidden
	}

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM categories WHERE category_id = $1`

	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting category %s", id)
	}

	return nil
}
