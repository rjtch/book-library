package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"expvar"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/zipkin"
	"github.com/book-library/cmd/book-api/internal/handlers"
	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/platform/database"
	"github.com/coreos/go-oidc/v3/oidc"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/pkg/errors"
	_ "github.com/rakyll/statik/fs"
	"github.com/spf13/viper"
	"go.opencensus.io/trace"
	"golang.org/x/oauth2"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"
var configPath = "./oidc/"
var configName = "config"
var configType = "json"

func main() {
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}

func run() error {
	ctx := context.Background()

	// =========================================================================
	// Logging

	logger := log.New(os.Stdout, "BOOKS : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	// read config files
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	err := viper.ReadInConfig()
	if err != nil {
		return errors.Wrap(err, "generating config usage failed")
	}
	// =========================================================================
	// Configuration
	oauth := auth.OAuthenticator{
		ClientID:       viper.GetString("oauth.clientID"),
		ClientSecret:   viper.GetString("oauth.clientSecret"),
		Endpoint:       viper.GetString("oauth.endpoint"),
		RedirectUrl:    viper.GetString("oauth.redirectUrl"),
		Issuer:         viper.GetString("oauth.issuer"),
		PublicKeyRS256: viper.GetString("oauth.publicKeyRS256"),
		Scopes:         viper.GetStringSlice("oauth.scopes"),
	}

	db := auth.DB{
		User:       viper.GetString("db.user"),
		Password:   viper.GetString("db.password"),
		Host:       viper.GetString("db.host"),
		Name:       viper.GetString("db.name"),
		DisableTLS: viper.GetBool("db.disabledTls"),
	}

	web := auth.Web{
		APIHost:         viper.GetString("web.apiHost"),
		DebugHost:       viper.GetString("web.debugHost"),
		ReadTimeout:     viper.GetDuration("web.readTimeout"),
		WriteTimeout:    viper.GetDuration("web.writeTimeout"),
		ShutdownTimeout: viper.GetDuration("web.shutdownTimeout"),
	}

	zipkinServer := auth.Zipkin{
		LocalEndpoint: viper.GetString("zipkin.localEndpoint"),
		ReporterURI:   viper.GetString("zipkin.reporterUri"),
		ServiceName:   viper.GetString("zipkin.serviceName"),
		Probability:   viper.GetFloat64("zipkin.probability"),
	}

	provider, err := oidc.NewProvider(ctx, oauth.Issuer)
	if err != nil {
		return errors.Wrap(err, "Provider could not been  found")
	}
	oidcConfig := &oidc.Config{
		ClientID:                   oauth.ClientID,
		InsecureSkipSignatureCheck: true,
	}
	verifier := provider.Verifier(oidcConfig)
	config := oauth2.Config{
		ClientID:     oauth.ClientID,
		ClientSecret: oauth.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  oauth.RedirectUrl,
		Scopes:       oauth.Scopes,
	}
	// config for the authenticator
	oauth.Config = config
	oauth.Provider = verifier
	// =========================================================================
	// App Starting

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(build)
	defer logger.Println("main : Completed")

	// =========================================================================
	// Initialize authentication support

	logger.Println("main : Started : Initializing authentication support")

	authenticator, err := auth.OAuthenticate(oauth)
	if err != nil {
		logger.Println("constructing authenticator has failed")
		return errors.Wrap(err, "constructing authenticator")
	} else {
		logger.Println("Verifier %s", verifier)
	}
	authenticator.Config = oauth.Config

	// =========================================================================
	// Start Database

	logger.Println("main : Started : Initializing database support")

	dbank, err := database.Open(database.Config{
		User:       db.User,
		Password:   db.Password,
		Host:       db.Host,
		Name:       db.Name,
		DisableTLS: db.DisableTLS,
	})
	if err != nil {
		logger.Println("connecting to db has failed")
		return errors.Wrap(err, "connecting to db")
	}

	defer func() {
		logger.Printf("main : Database Stopping : %s", db.Host)
		dbank.Close()
	}()

	// =========================================================================
	// Start Tracing Support

	logger.Println("main : Started : Initializing zipkin tracing support")

	localEndpoint, err := openzipkin.NewEndpoint(zipkinServer.ServiceName, zipkinServer.LocalEndpoint)
	if err != nil {
		return err
	}

	reporter := zipkinHTTP.NewReporter(zipkinServer.ReporterURI)
	ze := zipkin.NewExporter(reporter, localEndpoint)

	trace.RegisterExporter(ze)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.ProbabilitySampler(zipkinServer.Probability),
	})

	defer func() {
		logger.Printf("main : Tracing Stopping : %s", zipkinServer.LocalEndpoint)
		reporter.Close()
	}()

	// =========================================================================
	// Start Debug Service
	//
	// /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// /debug/vars - Added to the default mux by importing the expvar package.
	//
	// Not concerned with shutting this down when the application is shutdown.

	logger.Println("main : Started : Initializing debugging support")

	go func() {
		logger.Printf("main : Debug Listening %s", web.DebugHost)
		logger.Printf("main : Debug Listener closed : %v", http.ListenAndServe(web.DebugHost, http.DefaultServeMux))
	}()

	// =========================================================================
	// Start API Service

	logger.Println("main : Started : Initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         web.APIHost,
		Handler:      handlers.API(build, shutdown, logger, dbank, authenticator),
		ReadTimeout:  web.ReadTimeout,
		WriteTimeout: web.WriteTimeout,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		logger.Printf("main : API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:

		logger.Printf("main : %v : Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			logger.Printf("main : Graceful shutdown did not complete in %v : %v", web.ShutdownTimeout, err)
			err = api.Close()
		}

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGQUIT:
			return errors.New("integrity issue caused shutdown")
		case err != nil:
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}

	return nil
}
