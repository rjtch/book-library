package mid

import (
	"context"
	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/platform/web"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
	"net/http"
)

const (
	// default names for cookies and headers
	defaultJWTCookieName  = "SESSION-COOKIE"
	defaultXSRFCookieName = "XSRF-TOKEN"
)

//ErrForbidden is returned when a users doesn't have the required roles for doing an action
var ErrForbidden = web.NewRequestError(
	errors.New("you don't have the authorization for that action"),
	http.StatusForbidden,
)

//Authentication validates a jwt from the Authorization header
func Authentication(authenticator *auth.Authenticator) web.Middleware {

	//actual middleware to be execute
	f := func(after web.Handler) web.Handler {

		//wrapped handler around the next one
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Authentication")
			defer span.End()

			cookie, err := r.Cookie(defaultJWTCookieName)
			if err != nil {
				err := errors.New("expected session-cookie")
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			parts := cookie.Value
			claims, err := authenticator.ParseClaims(parts)
			if err != nil {
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			//TODO check if session-cookie is expired or if user is already logged out

			//Add claims to context so that they can be checked later on
			ctx = context.WithValue(ctx, auth.Key, claims)

			return after(ctx, w, r, params)
		}
		return h
	}
	return f
}

//HasRole checks and validate that an authenticated users has at least one of the required roles specified in the role's list
func HasRole(roles ...string) web.Middleware {
	// This is the actual middleware function to be executed.
	f := func(after web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.HasRole")
			defer span.End()

			claims, ok := ctx.Value(auth.Key).(auth.Claims)
			if !ok {

				return errors.New("claims missing from context: HasRole called without/before Authenticate")
			}

			if !claims.HasRole(roles...) {
				return ErrForbidden
			}

			return after(ctx, w, r, params)
		}

		return h
	}

	return f
}
