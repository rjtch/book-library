package loans

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	errors "github.com/pkg/errors"

	books "github.com/book-library/internal/books"
	auth "github.com/book-library/internal/platform/auth"
	"go.opencensus.io/trace"
)

const loansCollection = "loans"

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

//List retrieves a list of existing loans from the databse
func List(ctx context.Context, user auth.Claims, db *sqlx.DB) ([]Loan, error) {
	ctx, span := trace.StartSpan(ctx, "internal.loan.List")
	defer span.End()

	// If you do not have the admin role ...
	// then get outta here!
	//user, ok := ctx.Value(auth.Key).(auth.Claims)
	//if !ok {
	//	return nil, errors.New("claims missing from context")
	//}

	if !user.HasRole(auth.RoleAdmin) {
		return nil, ErrForbidden
	}

	loans := []Loan{}
	const q = `SELECT * FROM loans`

	if err := db.SelectContext(ctx, &loans, q); err != nil {
		return nil, errors.Wrap(err, "selecting loans")
	}

	return loans, nil
}

//InitNewLoan initiates a new loan when users want to loan a book
func InitNewLoan(ctx context.Context, user auth.Claims, n NewLoan, now time.Time, id string, db *sqlx.DB) (*Loan, error) {
	ctx, span := trace.StartSpan(ctx, "internal.loan.InitNewLoan")
	defer span.End()

	loan := Loan{
		ID:           uuid.New().String(),
		BookID:       id,
		BookISBN:     n.BookISBN,
		BookTitle:    n.BookTitle,
		BookQuantity: n.BookQuantity,
		LoanDate:     now.UTC(),
		ReturnDate:   now.Add(30).UTC(),
		UserID:       user.Subject,
	}

	const q = `INSERT INTO loans
	(loan_id, book_id, isbn, title, quantity, loan_date, date_return, user_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := db.ExecContext(
		ctx, q,
		loan.ID, loan.BookID, loan.BookISBN, loan.BookTitle, loan.BookQuantity,
		loan.LoanDate, loan.ReturnDate, user.Subject,
	)
	if err != nil {
		return nil, err
	}

	list, err := List(ctx, user, db)
	if err != nil {
		fmt.Printf("%v", list)
	}
	//get the book which is been lent
	book, err := books.Retrieve(ctx, loan.BookID, db)
	if err != nil {
		return nil, err
	}

	//reduce book quantity if loan succeeded
	book.Quantity = book.Quantity - 1
	nbk := books.UpdateBook{
		Quantity:    &book.Quantity,
		DateUpdated: &now,
	}

	//update book quantity
	books.Update(ctx, book.ID, nbk, now, user, db)

	bks2, errB2 := books.List(ctx, db)
	if errB2 != nil {
		fmt.Printf("%v", bks2)
	}
	return &loan, nil
}

//Retrieve retrieves a loan by id
func Retrieve(ctx context.Context, user auth.Claims, id string, db *sqlx.DB) (*Loan, error) {
	ctx, span := trace.StartSpan(ctx, "internal.loan.Retrieve")
	defer span.End()

	// If you do not have the required role or your not authorized ...
	// then get outta here!
	if !user.HasRole(auth.RoleAdmin) {
		return nil, ErrForbidden
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	//actual retrieven loan
	var loan Loan
	const q = `SELECT * FROM loans  WHERE loan_id  = $1`

	if err := db.GetContext(ctx, &loan, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting loan %q", id)
	}

	return &loan, nil
}

//EndUpALoan ends a loan after giving a book back
func EndUpALoan(ctx context.Context, user auth.Claims, now time.Time, id string, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.loan.EndUpALoan")
	defer span.End()

	// If you do not have the required role ...
	// then get outta here!
	if !user.HasRole(auth.RoleUser) {
		return ErrForbidden
	}

	loan, er := Retrieve(ctx, user, id, db)
	if er != nil {
		return ErrInvalidID
	}

	//get the book which is been lent
	book, errr := books.Retrieve(ctx, loan.BookID, db)
	if errr != nil {
		return ErrInvalidID
	}

	//update book quantity if laon succeeded
	book.Quantity++
	nbk := books.UpdateBook{
		Quantity:    &book.Quantity,
		DateUpdated: &now,
	}

	//update book quantity
	books.Update(ctx, id, nbk, now, user, db)
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM loans WHERE loan_id = $1`

	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting loan %s", id)
	}

	return nil
}

// Update replaces a Loan document in the database.
func Update(ctx context.Context, id string, upd UpdateLoan, now time.Time, user auth.Claims, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "internal.Loan.Update")
	defer span.End()

	// // If you do not have the admin role ...
	// // then get outta here!
	//if !user.HasRole(auth.RoleUser) && user.Subject != id {
	//	return ErrForbidden
	//}

	if !user.HasRole(auth.RoleAdmin) {
		return ErrForbidden
	}

	loan, err := Retrieve(ctx, user, id, db)
	if err != nil {
		return err
	}

	if upd.BookQuantity != nil {
		loan.BookQuantity += *upd.BookQuantity
	}

	if upd.ReturnDate != nil {
		loan.ReturnDate = *upd.ReturnDate
	}


	const q = `UPDATE loans SET
		"isbn" = $2,
		"quantity" = $3,
		"date_return" = $4
		WHERE loan_id = $1`
	_, err = db.ExecContext(ctx, q, id,
		upd.BookISBN, upd.BookQuantity, upd.ReturnDate,
	)
	if err != nil {
		return errors.Wrap(err, "updating users")
	}

	return nil
}
