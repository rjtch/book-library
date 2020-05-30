package handlers

import (
	"context"
	"net/http"

	category "github.com/book-library/internal/book-category"
	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/platform/web"
	"github.com/book-library/internal/users"
	"github.com/jmoiron/sqlx"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
)

//BookCategory represents the BookCategory API method handler set.
type BookCategory struct {
	db *sqlx.DB
}

//List returns all the existing Bookcategories from the system to the world
func (c *BookCategory) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.book-category.List")
	defer span.End()

	categories, err := category.List(ctx, c.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, categories, http.StatusOK)
}

//Create creates a new bookCategory into the system
func (c *BookCategory) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.book-category.Create")
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

	var nb category.NewBookCategory
	if err := web.Decode(r, &nb); err != nil {
		return errors.Wrap(err, "creation a new book category")
	}

	bkCategory, err := category.Create(ctx, v.Now, nb, claims, c.db)
	if err != nil {
		return errors.Wrapf(err, "Category: %+v", &bkCategory)
	}
	return web.Respond(ctx, w, bkCategory, http.StatusCreated)

}

//Update updates a specified bookCategory in the database
func (c *BookCategory) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.book-category.Update")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return errors.New("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	var udp category.UpdateBookCategory
	if err := web.Decode(r, &udp); err != nil {
		return errors.Wrap(err, "")
	}

	err := category.Update(ctx, params["id"], udp, v.Now, claims, c.db)
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

//Delete deletes a unique bookCategory from the database
func (c *BookCategory) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.book-category.Delete")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	if claims.HasRole(auth.RoleAdmin) {
		return errors.New("you don't have role to execute this action")
	}

	err := category.Delete(ctx, params["id"], claims, c.db)
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

//Retreive returns the value of a specified users from the system to the world
func (c *BookCategory) Retreive(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.Retrieve")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	cat, err := category.Retrieve(ctx, claims, c.db, params["id"])
	if err != nil {
		switch err {
		case category.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		case category.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case category.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}
	return web.Respond(ctx, w, cat, http.StatusOK)
}
