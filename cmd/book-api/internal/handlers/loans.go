package handlers

import (
	"context"
	"net/http"

	loans "github.com/book-library/internal/loan"
	"github.com/jmoiron/sqlx"

	"github.com/book-library/internal/books"
	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/platform/web"
	"github.com/book-library/internal/users"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
)

//Loan represents the Loan API method handler set.
type Loan struct {
	db *sqlx.DB
}

//List returns all the existing Loan from the system to the world
func (l *Loan) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.loans.List")
	defer span.End()

	//get user_id from the url
	id := params["user_id"]

	//check if token does already exist
	ok, err := users.IsLoggedOut(ctx, l.db, id, r.Header.Get("bearer"))
	if !ok {
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	allLoans := []loans.Loan{};

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	loans, err := loans.List(ctx, claims, l.db)
	if err != nil {
		return err
	}

	if len(loans) != 0 {
		for _, l := range loans {
			if l.UserID == claims.StandardClaims.Subject {
				allLoans = append(allLoans, l)
			}
		}
	} else {
		return errors.Wrap(nil, "your are not allow to execute this action")
	}
	return web.Respond(ctx, w, allLoans, http.StatusOK)
}

//Retrieve returns the value of a specified Loan from the system to the world
func (l *Loan) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.loans.Retrieve")
	defer span.End()

	//get user_id from the url
	id := params["user_id"]

	//check if token does already exist
	ok, err := users.IsLoggedOut(ctx, l.db, id, r.Header.Get("bearer"))
	if !ok {
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	loan, err := loans.Retrieve(ctx, claims, params["id"], l.db, claims.Subject)
	if err != nil {
		switch err {
		case books.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		case books.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case books.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}
	return web.Respond(ctx, w, loan, http.StatusOK)
}

//Create creates a new Loan into the system
func (l *Loan) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.loans.Create")
	defer span.End()

	//get user_id from the url
	id := params["user_id"]

	//check if token does already exist
	ok, err := users.IsLoggedOut(ctx, l.db, id, r.Header.Get("bearer"))
	if !ok {
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	//we retreive hier as claim the Value(state of each request) because we are in this case creating a new users
	//so he doesn't have any claim and role yet and have to be created first thats why a keyValue from the web
	//is used instead
	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("web value missing from context")
	}

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	//TODO check always do this thing in repository layer
	book, err := books.Retrieve(ctx, params["id"], l.db)
	if err != nil {
		return errors.Wrapf(err, "Book: %+v", &book)
	}

	var nl loans.NewLoan
	if err := web.Decode(r, &nl); err != nil {
		return errors.Wrap(err, "Error when decoding the request's body")
	}

	loan, err := loans.InitNewLoan(ctx, claims, nl, v.Now, book.ID, l.db)
	if err != nil {
		switch err {
		case users.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		case users.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case users.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}
	return web.Respond(ctx, w, loan, http.StatusCreated)
}

//Update updates a specified Loan in the database
func (l *Loan) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.loans.Update")
	defer span.End()

	//get user_id from the url
	id := params["user_id"]

	//check if token does already exist
	ok, err := users.IsLoggedOut(ctx, l.db, id, r.Header.Get("bearer"))
	if !ok {
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return errors.New("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	var udl loans.UpdateLoan
	if err := web.Decode(r, &udl); err != nil {
		return errors.Wrap(err, "")
	}

	loan, err := loans.Retrieve(ctx, claims, params["id"], l.db, claims.Subject)
	if err != nil {
		return errors.New("you don't have wright to execute this action")
	}

	err = loans.Update(ctx, string(loan.ID), udl, v.Now, claims, l.db)
	if err != nil {
		switch err {
		case users.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		case users.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case users.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, loan, http.StatusOK)
}

//Delete deletes a unique Loan from the database
func (l *Loan) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.loans.Delete")
	defer span.End()

	//get user_id from the url
	id := params["user_id"]

	//check if token does already exist
	ok, err := users.IsLoggedOut(ctx, l.db, id, r.Header.Get("bearer"))
	if !ok {
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return errors.New("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	err = loans.EndUpALoan(ctx, claims, v.Now, params["id"], l.db)
	if err != nil {
		switch err {
		case users.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		case users.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case users.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusOK)
}
