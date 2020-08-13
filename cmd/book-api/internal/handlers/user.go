package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/platform/web"
	"github.com/book-library/internal/users"
	"github.com/jmoiron/sqlx"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
)

//User represents the Users API method handler set.
type User struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
}

//List returns all the existing users from the system to the world
func (u *User) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.List")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		if !claims.HasRole(auth.RoleAdmin) {
			return errors.New("claims missing from context")
		}
	}

	usr, err := users.List(ctx, claims, u.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

//Retrieve returns the value of a specified users from the system to the world
func (u *User) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.Retrieve")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	user, err := users.Retrieve(ctx, claims, u.db, params["id"])
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
	return web.Respond(ctx, w, user, http.StatusOK)
}

//Retrieve returns the value of a specified users from the system to the world
func (u *User) RetrieveMe(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.Retrieve")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	user, err := users.RetrieveMe(ctx, claims, u.db)
	if err != nil {
		switch err {
		case users.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		case users.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case users.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", nil)
		}
	}
	return web.Respond(ctx, w, user, http.StatusOK)
}

//Create creates a new users into the system
func (u *User) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.Create")
	defer span.End()

	//we retreive hier as claim the Value(state of each request) because we are in this case creating a new users
	//so he doesn't have any claim and role yet and have to be created first thats why a keyValue from the web
	//is used instead
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return errors.New("web value missing from context")
	}

	var nu users.NewUser
	if err := web.Decode(r, &nu); err != nil {
		return errors.Wrap(err, "")
	}

	user, err := users.Create(ctx, u.db, nu, v.Now)
	if err != nil {
		return errors.Wrapf(err, "User: %+v", &user)
	}
	return web.Respond(ctx, w, user, http.StatusCreated)

}

//Update updates a specified users in the database
func (u *User) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.Update")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return errors.New("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	var udp users.UpdateUser
	if err := web.Decode(r, &udp); err != nil {
		return errors.Wrap(err, "")
	}

	err := users.Update(ctx, claims, u.db, params["id"], udp, v.Now)
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

//Delete deletes a unique users from the database
func (u *User) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.Delete")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	if claims.HasRole(auth.RoleAdmin) {
		return errors.New("you don't have role to execute this action")
	}

	err := users.Delete(ctx, u.db, params["id"])
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

//TokenAuthenticator handles request to authenticate the users and expects a request using Basic Auth with the User's email
//and password. It reponds with a jwt
func (u *User) TokenAuthenticator(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.TokenAuthenticator")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	// BasicAuth returns the username and password provided in the request's
	// Authorization header, if the request uses HTTP Basic Authentication.
	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide valide email and password for authentication")
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := users.Authenticate(ctx, u.db, v.Now, email, pass)

	fmt.Println("EMAIL ", email)
	fmt.Println("PASSWORD ", pass)
	fmt.Println("ERRRROORRR ", err)
	fmt.Println("CLAIMS ", claims)

	if err != nil {
		switch err {
		case users.ErrAuthenticationFailure:
			return web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return errors.Wrap(err, "authenticating")
		}
	}

	var tk struct {
		Token string `json:"token"`
	}

	tk.Token, err = u.authenticator.GenerateToken(claims)

	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	return web.Respond(ctx, w, tk, http.StatusOK)
}
