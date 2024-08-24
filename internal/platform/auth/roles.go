package auth

import (
	"fmt"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

// These are the expected values for Claims.Roles.
const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

type Role struct {
	RoleAdmin []string
	RoleUser  []string
}

// ctxKey represents the type of value for the context key.
type ctxKey int

// Key is used to store/retrieve a Claims value from a context.Context.
const Key ctxKey = 1

// Claims represents the authorization claims transmitted via a JWT.
type Claims struct {
	Roles          []string  `json:"roles"`
	StandardClaims jwt.Token `json:"standardClaims"`
	Csrf           string    `json:"csrf"`
}

// NewClaims constructs a Claims value for the identified users. The Claims
// expire within a specified duration of the provided time. Additional fields
// of the Claims can be set after calling NewClaims is desired.
func NewClaims(token jwt.Token, roles []string, csrf string) Claims {
	c := Claims{
		Roles:          roles,
		Csrf:           csrf,
		StandardClaims: token,
	}

	return c
}

// Valid is called during the parsing of a token.
func (c Claims) Valid() error {
	for _, r := range c.Roles {
		switch r {
		case RoleAdmin, RoleUser: // Role is valid.
		default:
			return fmt.Errorf("invalid role %q", r)
		}
	}

	t, err := c.StandardClaims.Claims.GetExpirationTime()
	if err != nil {
		return errors.Wrap(err, "validating standard claims")
	}

	if t.Time.IsZero() {
		return errors.New("Token already expired")
	}

	return nil
}

// HasRole returns true if the claims has at least one of the provided roles.
func (c Claims) HasRole(roles ...string) bool {
	for _, has := range c.Roles {
		for _, want := range roles {
			if has == want {
				return true
			}
		}
	}
	return false
}
