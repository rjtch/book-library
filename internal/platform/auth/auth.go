package auth

import (
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
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
	ClientID       string
	ClientSecret   string
	Endpoint       string
	RedirectUrl    string
	Issuer         string
	PublicKeyRS256 string
	Scopes         []string
	Config         oauth2.Config
	Provider       *oidc.IDTokenVerifier
}

func OAuthenticate(auth OAuthenticator) (*OAuthenticator, error) {
	if auth.ClientID == "" {
		return nil, ErrClientIDError
	}

	if auth.ClientSecret == "" {
		return nil, ErrClientSecretError
	}

	if auth.Endpoint == "" {
		return nil, ErrEndpointError
	}

	if auth.RedirectUrl == "" {
		return nil, ErrRedirectUrlError
	}

	if auth.Issuer == "" {
		return nil, ErrIssuerError
	}

	if len(auth.Scopes) == 0 {
		return nil, ErrScopesError
	}

	oauth := OAuthenticator{
		Config:         auth.Config,
		Provider:       auth.Provider,
		PublicKeyRS256: auth.PublicKeyRS256,
		Issuer:         auth.Issuer,
		Scopes:         auth.Scopes,
		RedirectUrl:    auth.RedirectUrl,
		Endpoint:       auth.Endpoint,
		ClientID:       auth.ClientID,
		ClientSecret:   auth.ClientSecret,
	}

	return &oauth, nil
}
