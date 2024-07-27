package auth

import (
	"time"

	errors "github.com/pkg/errors"
)

// Predefined errors identify expected failure conditions.
var (
	//ErrKIDFormatError is used when kid format is not a string
	ErrClientIDError = errors.New("Client-id (Cid) is not provided")

	//ErrKIDFormatError is used when kid format is not a string
	ErrClientSecretError = errors.New("client-secret is not provided")

	//ErrKIDFormatError is used when kid format is not a string
	ErrRedirectUrlError = errors.New("redirect-url is not provided")

	//ErrKIDFormatError is used when kid format is not a string
	ErrScopesError = errors.New("scopes are missing or malformed")

	//ErrKIDFormatError is used when kid format is not a string
	ErrEndpointError = errors.New("endpoint is not provided")
	ErrIssuerError   = errors.New("issuer is not provided")
)

type Web struct {
	APIHost         string        `json:"apiHost,omitempty"`
	DebugHost       string        `json:"debugHost,omitempty"`
	ReadTimeout     time.Duration `json:"readTimeout,omitempty"`
	WriteTimeout    time.Duration `json:"writeTimeout,omitempty"`
	ShutdownTimeout time.Duration `json:"shutdownTimeout,omitempty"`
}
type DB struct {
	User       string `json:"user,omitempty"`
	Password   string `json:"password,omitempty"`
	Host       string `json:"host,omitempty"`
	Name       string `json:"name,omitempty"`
	DisableTLS bool   `json:"disableTLS,omitempty"`
}
type Auth struct {
	KeyID string `json:"keyid,omitempty"`
	//			PrivateKeyFile string `conf:"default:/app-library/private.pem"`
	PrivateKeyFile string `json:"privateKeyFile,omitempty"`
	Algorithm      string `json:"algorithm,omitempty"`
}
type OAuth struct {
	ClientID     string   `json:"clientID,omitempty"`
	ClientSecret string   `json:"clientSecret,omitempty"`
	Endpoint     string   `json:"endpoint,omitempty"`
	RedirectUrl  string   `json:"redirectUrl,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	Issuer       string   `json:"issuer,omitempty"`
}
type Zipkin struct {
	LocalEndpoint string  `json:"localEndpoint,omitempty"`
	ReporterURI   string  `json:"reporterURI,omitempty"`
	ServiceName   string  `json:"serviceName,omitempty"`
	Probability   float64 `json:"probability,omitempty"`
}

type OAuthenticator struct {
	ClientID     string
	ClientSecret string
	Endpoint     string
	RedirectUrl  string
	Issuer       string
	Scopes       []string
}

func OAuthenticate(clientId string, secret string, endpoint string, redirect string, issuer string, scopes []string) (*OAuthenticator, error) {
	if clientId == "" {
		return nil, ErrClientIDError
	}

	if secret == "" {
		return nil, ErrClientSecretError
	}

	if endpoint == "" {
		return nil, ErrEndpointError
	}

	if redirect == "" {
		return nil, ErrRedirectUrlError
	}

	if issuer == "" {
		return nil, ErrIssuerError
	}

	if len(scopes) == 0 {
		return nil, ErrScopesError
	}

	auth := OAuthenticator{
		ClientID:     clientId,
		ClientSecret: secret,
		Endpoint:     endpoint,
		RedirectUrl:  redirect,
		Issuer:       issuer,
		Scopes:       scopes,
	}

	return &auth, nil
}
