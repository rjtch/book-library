package mid

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"log"
	"net/http"
	"slices"
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
// TODO extend this methode with role.
func Authentication(authenticator *auth.OAuthenticator, role string) web.Middleware {

	//actual middleware to be execute
	f := func(after web.Handler) web.Handler {

		//wrapped handler around the next one
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Authentication")
			defer span.End()
			stringToken := r.Header.Get(authorization)
			// Parse the authorization header.
			parts := strings.Split(stringToken, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != bearer {
				return errors.New("expected authorization header format: bearer <token>")
			}

			err, token := extractClaims(w, r, ctx, authenticator, role)
			if err != nil {
				return errors.New(" authorization header token bearer not valid")
			}

			//Add claims to context so that they can be checked later on
			ctx = context.WithValue(ctx, auth.Key, token)

			return after(ctx, w, r, params)
		}
		return h
	}
	return f
}

func extractClaims(w http.ResponseWriter, request *http.Request, ctx context.Context, oauth *auth.OAuthenticator, role string) (error, *jwt.Token) {
	stringToken := request.Header.Get(authorization)
	secretKey, err := parseRSAPublicKey(oauth.PublicKeyRS256)
	if err != nil {
		return errors.New("Cannot load certificate: " + err.Error()), nil
	}

	// Parse the authorization header.
	parts := strings.Split(stringToken, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != bearer {
		return errors.New("expected authorization header format: bearer <token>"), nil
	}

	_, err = oauth.Provider.Verify(ctx, parts[1])
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return errors.New("Failed to verify ID Token " + err.Error()), nil
	}

	token, err := jwt.ParseWithClaims(parts[1], jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("there's an error with the signing method " + err.Error())
		}
		return secretKey, nil
	})

	claims, ok := ctx.Value(auth.Key).(*jwt.Token)
	if !ok {
		return errors.New("claims missing from context"), nil
	}
	roles := ParseRealmRoles(claims.Claims.(jwt.MapClaims))
	if len(roles) == 0 {
		return errors.New("Not roles related to the this user"), nil
	}
	if !slices.Contains(roles, role) {
		return errors.New("You don't have the wright to see the required resources."), nil
	}
	if errors.Is(err, jwt.ErrSignatureInvalid) {
		return errors.New(err.Error()), nil
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Println("token is valid")
	}

	return nil, token
}

func parseRSAPublicKey(base64Str string) (*rsa.PublicKey, error) {
	buf, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, err
	}
	parsedKey, err := x509.ParsePKIXPublicKey(buf)
	if err != nil {
		return nil, err
	}
	publicKey, ok := parsedKey.(*rsa.PublicKey)
	if ok {
		return publicKey, nil
	}
	return nil, errors.New("unexpected key type for public key")
}

func ParseRealmRoles(claims jwt.MapClaims) []string {
	var realmRoles []string = make([]string, 0)

	if claim, ok := claims["realm_access"]; ok {
		if roles, ok := claim.(map[string]interface{})["roles"]; ok {
			for _, role := range roles.([]interface{}) {
				realmRoles = append(realmRoles, role.(string))
			}
		}
	}
	return realmRoles
}
