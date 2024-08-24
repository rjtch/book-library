package mid

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/platform/web"
	"github.com/golang-jwt/jwt/v5"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const (
	// default names for cookies and headers
	defaultJWTCookieName  = "session-cookie"
	defaultXSRFCookieName = "x-xsrf-token"
	authorization         = "authorization"
	bearer                = "bearer"
)

// ErrForbidden is returned when a users doesn't have the required roles for doing an action
var ErrForbidden = web.NewRequestError(
	errors.New("you don't have the authorization for that action"),
	http.StatusForbidden,
)

// Authentication validates a jwt and the csrf cookie from the Authorization header
func Authentication(authenticator *auth.OAuthenticator) web.Middleware {

	//actual middleware to be execute
	f := func(after web.Handler) web.Handler {

		//wrapped handler around the next one
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Authentication")
			defer span.End()

			// Expecting: bearer <token>
			authStr, err := extractClaims(w, r, authenticator.PubKey)
			if err != nil {
				return errors.New("expected authorization header format: bearer <token>")
			}

			//Add claims to context so that they can be checked later on
			ctx = context.WithValue(ctx, auth.Key, authStr)

			return after(ctx, w, r, params)
		}
		return h
	}
	return f
}

func extractClaims(_ http.ResponseWriter, request *http.Request, pubkey string) (error, *jwt.Token) {
	stringToken := request.Header.Get(authorization)
	// Parse the authorization header.
	parts := strings.Split(stringToken, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != bearer {
		return errors.New("expected authorization header format: bearer <token>"), nil
	}

	// Create a file for the public key information in PEM form.
	publicFile, err := os.Create("public.pem")
	if err != nil {
		return errors.New("creating public file: "), nil
	}
	defer publicFile.Close()

	pubKey, err := os.ReadFile("public.pem")
	if err != nil {
		return errors.New("error public file: "), nil
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pubKey))
	if err != nil {
		return errors.New("failed to parse pubkey"), nil
	}

	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("there's an error with the signing method")
		}
		return key, nil
	})

	if err != nil {
		return err, nil
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Println("token is valid")
	}

	return nil, token
}
