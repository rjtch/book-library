package handlers

import (
	"context"
	"net/http"

	"github.com/book-library/internal/books"
	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/platform/web"
	"github.com/book-library/internal/users"
	"github.com/jmoiron/sqlx"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
)

//Book represents the Books API method handler set.
type Book struct {
	db *sqlx.DB
}

//List returns all the existing Book from the system to the world
func (b *Book) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.books.List")
	defer span.End()

	books, err := books.List(ctx, b.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, books, http.StatusOK)
}

//Retrieve returns the value of a specified Book from the system to the world
func (b *Book) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.books.Retrieve")
	defer span.End()

	book, err := books.Retrieve(ctx, params["id"], b.db)
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
	return web.Respond(ctx, w, book, http.StatusOK)
}

//Create creates a new Book into the system
func (b *Book) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.books.Create")
	defer span.End()

	//we retreive hier as claim the Value(state of each request) because we are in this case creating a new users
	//so he doesn't have any claim and role yet and have to be created first thats why a keyValue from the web
	//is used instead
	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	if !claims.HasRole(auth.RoleAdmin) {
		return errors.New("you don't have role to execute this action")
	}

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var nb books.NewBook
	if err := web.Decode(r, &nb); err != nil {
		return errors.Wrap(err, "")
	}

	book, err := books.Create(ctx, v.Now, nb, claims, b.db)
	if err != nil {
		return errors.Wrapf(err, "Book: %+v", &book)
	}
	return web.Respond(ctx, w, book, http.StatusCreated)

}

//Update updates a specified Book in the database
func (b *Book) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.books.Update")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return errors.New("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	if !claims.HasRole(auth.RoleAdmin) {
		return errors.New("not authorized to execute this action")
	}

	var udp books.UpdateBook
	if err := web.Decode(r, &udp); err != nil {
		return errors.Wrap(err, "")
	}

	err := books.Update(ctx, params["id"], udp, v.Now, claims, b.db)
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

//Delete deletes a unique Book from the database
func (b *Book) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.books.Delete")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	if !claims.HasRole(auth.RoleAdmin) {
		return errors.New("you don't have role to execute this action")
	}

	err := books.Delete(ctx, params["id"], claims, b.db)
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
