package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/platform/web"
	"github.com/book-library/internal/users"
	"github.com/jmoiron/sqlx"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const (
	AllowOriginKey      string = "Access-Control-Allow-Origin"
	AllowCredentialsKey string = "Access-Control-Allow-Credentials"
	AllowHeadersKey     string = "Access-Control-Allow-Headers"
	// default names for cookies and headers
	defaultJWTCookieName string = "session-cookie"
	defaultXsrfToken     string = "x-xsrf-token"
	OriginKey            string = "Origin"
)

// User represents the Users API method handler set.
type User struct {
	Db            *sqlx.DB
	authenticator *auth.OAuthenticator
}

// List returns all the existing users from the system to the world
func (u *User) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.List")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		if !claims.HasRole(auth.RoleAdmin) {
			return errors.New("claims missing from context")
		}
	}

	usr, err := users.List(ctx, u.Db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

// Retrieve returns the value of a specified users from the system to the world
func (u *User) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.Retrieve")
	defer span.End()

	user, err := users.Retrieve(ctx, u.Db, params["id"])
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

// Create creates a new users into the system
func (u *User) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.Create")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		if !claims.HasRole(auth.RoleAdmin) {
			return errors.New("claims missing from context")
		}
	}

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

	user, err := users.Create(ctx, u.Db, nu, v.Now)
	if err != nil {
		return errors.Wrapf(err, "User: %+v", &user)
	}
	return web.Respond(ctx, w, user, http.StatusCreated)

}

// Update updates a specified users in the database
func (u *User) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.Update")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return errors.New("web value missing from context")
	}

	var udp users.UpdateUser
	if err := web.Decode(r, &udp); err != nil {
		return errors.Wrap(err, "")
	}

	err := users.Update(ctx, u.Db, params["id"], udp, v.Now)
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

// Delete deletes a unique users from the database
func (u *User) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.Delete")
	defer span.End()

	err := users.Delete(ctx, u.Db, params["id"])
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

// TokenAuthenticator handles request to authenticate the users and expects a request using Basic Auth with the User's email
// and password. It responds with a jwt
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
		err := errors.New("must provide valid email and password for authentication")
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := users.Authenticate(ctx, u.Db, v.Now, email, pass)
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

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")
	//set the hidden x-xsrf-token header
	w.Header().Add(defaultXsrfToken, claims.Csrf)

	enableCors(&w)

	return web.Respond(ctx, w, tk.Token, http.StatusOK)
}

// RefreshToken refreshes a given claims by issuing a new token
func (u *User) RefreshToken(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.RefreshToken")
	defer span.End()

	_, ok := ctx.Value(web.KeyValues).(*web.Values)
	if ok {
		return web.NewShutdownError("web value missing from context")
	}

	cookie, _ := r.Cookie(defaultJWTCookieName)
	if cookie.Value == "" {
		err := errors.New("expected session-cookie")
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	//add cookies back int the header
	http.SetCookie(w, cookie)
	return web.Respond(ctx, w, "Refreshed token", http.StatusOK)
}

func (u *User) Logout(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.users.Logout")
	defer span.End()

	_, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	//get user_id from the url
	id := params["user_id"]
	err := users.Logout(ctx, u.Db, id)
	if err != nil {
		return errors.Wrap(err, "could not logout cookie already expired")
	}

	//get actual cookie from request
	cookie, _ := r.Cookie(defaultJWTCookieName)
	if cookie.Value == "" {
		err := errors.New("expected session-cookie")
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	//invalidate cookies after session is deleted from the db
	cookie.MaxAge = 0
	cookie.Expires = time.Now()

	//send the invalidated cookie back to the client
	http.SetCookie(w, cookie)

	return web.Respond(ctx, w, "logout was successful", http.StatusOK)
}

// enableCors enables cross origin control
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set(AllowOriginKey, "*")
	(*w).Header().Set(AllowCredentialsKey, "*")
	(*w).Header().Set(AllowHeadersKey, "*")
	(*w).Header().Set(OriginKey, "*")
}
