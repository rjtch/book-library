package main

import (
	"context"
	"expvar"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

func run() error {
	ctx := context.Background()

	// =========================================================================
	// Logging

	log := log.New(os.Stdout, "BOOKS : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	// read config files
	//	viper.SetConfigFile(configFile)
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	err := viper.ReadInConfig()
	if err != nil {
		return errors.Wrap(err, "generating config usage failed")
	}
	scopes := make([]string, 6)
	scopes = append(scopes, viper.GetString("oauth.scopes"))
	// =========================================================================
	// Configuration
	oauth := auth.OAuthenticator{
		ClientID:     viper.GetString("oauth.clientID"),
		ClientSecret: viper.GetString("oauth.clientSecret"),
		Endpoint:     viper.GetString("oauth.endpoint"),
		RedirectUrl:  viper.GetString("oauth.redirectUrl"),
		Issuer:       viper.GetString("oauth.issuer"),
		Scopes:       scopes,
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

	provider := oidc.InsecureIssuerURLContext(ctx, viper.GetString("oauth.issuer"))
	log.Printf("main : provider context version %q", provider)
	// =========================================================================
	// App Starting

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(build)
	defer log.Println("main : Completed")

	// =========================================================================
	// Initialize authentication support

	log.Println("main : Started : Initializing authentication support")

	authenticator, err := auth.OAuthenticate(oauth.ClientID, oauth.ClientSecret, oauth.Endpoint, oauth.RedirectUrl, oauth.Issuer, oauth.Scopes)
	if err != nil {
		return errors.Wrap(err, "constructing authenticator")
	}

	// =========================================================================
	// Start Database

	log.Println("main : Started : Initializing database support")

	dbank, err := database.Open(database.Config{
		User:       db.User,
		Password:   db.Password,
		Host:       db.Host,
		Name:       db.Name,
		DisableTLS: db.DisableTLS,
	})
	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}

	defer func() {
		log.Printf("main : Database Stopping : %s", db.Host)
		dbank.Close()
	}()

	// =========================================================================
	// Start Tracing Support

	log.Println("main : Started : Initializing zipkin tracing support")

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
		log.Printf("main : Tracing Stopping : %s", zipkinServer.LocalEndpoint)
		reporter.Close()
	}()

	// =========================================================================
	// Start Debug Service
	//
	// /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// /debug/vars - Added to the default mux by importing the expvar package.
	//
	// Not concerned with shutting this down when the application is shutdown.

	log.Println("main : Started : Initializing debugging support")

	go func() {
		log.Printf("main : Debug Listening %s", web.DebugHost)
		log.Printf("main : Debug Listener closed : %v", http.ListenAndServe(web.DebugHost, http.DefaultServeMux))
	}()

	// =========================================================================
	// Start API Service

	log.Println("main : Started : Initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         web.APIHost,
		Handler:      handlers.API(build, shutdown, log, dbank, authenticator),
		ReadTimeout:  web.ReadTimeout,
		WriteTimeout: web.WriteTimeout,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main : API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:

		log.Printf("main : %v : Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main : Graceful shutdown did not complete in %v : %v", web.ShutdownTimeout, err)
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
