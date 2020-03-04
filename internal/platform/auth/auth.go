package auth

import (
	"crypto/rsa"

	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	errors "github.com/pkg/errors"
)

// Predefined errors identify expected failure conditions.
var (
	// ErrPrivateKeyNil cannot be nil.
	ErrPrivateKeyNil = errors.New("private key cannot be nil")

	// ErrBlankActiveID is used when acitveID is blank.
	ErrBlankActiveID = errors.New("active kid cannot be blank")

	// ErrPublicKeyNil is used when public key is nil
	ErrPublicKeyNil = errors.New("public key function cannot be nil")

	//ErrKIDMiss is used when kid is missing
	ErrKIDMiss = errors.New("missing key id (kid) in token header")

	//ErrKIDFormatError is used when kid format is not a string
	ErrKIDFormatError = errors.New("users token key id (kid) must be string")
)

// KeyLookupFunc is used to map a JWT key id (kid) to the corresponding public key.
// It is a requirement for creating an Authenticator.
//
// * Private keys should be rotated. During the transition period, tokens
// signed with the old and new keys can coexist by looking up the correct
// public key by key id (kid).
//
// * Key-id-to-public-key resolution is usually accomplished via a public JWKS
// endpoint. See https://auth0.com/docs/jwks for more details.
type KeyLookupFunc func(kid string) (*rsa.PublicKey, error)

// NewSimpleKeyLookupFunc is a simple implementation of KeyFunc that only ever
// supports one key. This is easy for development but in production should be
// replaced with a caching layer that calls a JWKS endpoint.
func NewSimpleKeyLookupFunc(activeKID string, publicKey *rsa.PublicKey) KeyLookupFunc {
	f := func(kid string) (*rsa.PublicKey, error) {
		if activeKID != kid {
			return nil, fmt.Errorf("unrecognized key id %q", kid)
		}
		return publicKey, nil
	}
	return f
}

// Authenticator is used to authenticate clients. It can generate a token for a
// set of users claims and recreate the claims by parsing the token.
type Authenticator struct {
	privateKey       *rsa.PrivateKey
	activeID         string
	algorithm        string
	pubKeyLookupFunc KeyLookupFunc
	parser           *jwt.Parser
}

// NewAuthenticator creates an *Authenticator for use. It will error if:
// - The private key is nil.
// - The public key func is nil.
// - The key ID is blank.
// - The specified algorithm is unsupported.
func NewAuthenticator(privateKey *rsa.PrivateKey, activeKID, algorithm string, publicKeyLookupFunc KeyLookupFunc) (*Authenticator, error) {
	if privateKey == nil {
		return nil, ErrPrivateKeyNil
	}

	if activeKID == "" {
		return nil, ErrBlankActiveID
	}

	if publicKeyLookupFunc == nil {
		return nil, ErrPublicKeyNil
	}

	parser := jwt.Parser{
		ValidMethods: []string{algorithm},
	}

	a := Authenticator{
		privateKey:       privateKey,
		activeID:         activeKID,
		algorithm:        algorithm,
		pubKeyLookupFunc: publicKeyLookupFunc,
		parser:           &parser,
	}
	return &a, nil
}

//GenerateToken generates a signed jwt token string representing users claims
func (a *Authenticator) GenerateToken(claims Claims) (string, error) {
	method := jwt.GetSigningMethod(a.algorithm)

	tk := jwt.NewWithClaims(method, claims)
	tk.Header["kid"] = a.activeID

	str, err := tk.SignedString(a.privateKey)
	if err != nil {
		return "", errors.Wrap(err, "signing token")
	}
	return str, nil
}

// ParseClaims recreates the Claims that were used to generate a token. It
// verifies that the token was signed using our key.
func (a *Authenticator) ParseClaims(tokenStr string) (Claims, error) {

	// f is a function that returns the public key for validating a token. We use
	// the parsed (but unverified) token to find the key id. That ID is passed to
	// our KeyFunc to find the public key to use for verification.
	keyFunc := func(tk *jwt.Token) (interface{}, error) {
		kid, ok := tk.Header["kid"]
		if !ok {
			return nil, ErrKIDMiss
		}

		userKID, ok := kid.(string)
		if !ok {
			return nil, ErrKIDFormatError
		}
		return a.pubKeyLookupFunc(userKID)
	}

	var claim Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claim, keyFunc)
	if err != nil {
		return Claims{}, errors.Wrap(err, "parsing token")
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return claim, nil
}
