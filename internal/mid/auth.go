package mid

import (
	"context"
	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/platform/web"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
	"net/http"
	"strings"
)

const (
	// default names for cookies and headers
	defaultJWTCookieName  = "session-cookie"
	defaultXSRFCookieName = "x-xsrf-token"
)

//ErrForbidden is returned when a users doesn't have the required roles for doing an action
var ErrForbidden = web.NewRequestError(
	errors.New("you don't have the authorization for that action"),
	http.StatusForbidden,
)

//Authentication validates a jwt and the csrf cookie from the Authorization header
func Authentication(authenticator *auth.Authenticator) web.Middleware {

	//actual middleware to be execute
	f := func(after web.Handler) web.Handler {

		//wrapped handler around the next one
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Authentication")
			defer span.End()

			// Expecting: bearer <token>
			authStr := r.Header.Get("authorization")

			// Parse the authorization header.
			parts := strings.Split(authStr, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return errors.New("expected authorization header format: bearer <token>")
			}

			// Validate the token is signed by us.
			claims, err := authenticator.ParseClaims(parts[1])
			if err != nil {
				return errors.New("Token does not exist")
			}

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
