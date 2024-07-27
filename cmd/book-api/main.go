package main

import (
	"context"
	"expvar"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/zipkin"
	"github.com/ardanlabs/conf"
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
	// =========================================================================
	// Configuration

	var cfg struct {
		Web struct {
			APIHost         string        `json:"apiHost,omitempty"`
			DebugHost       string        `json:"debugHost,omitempty"`
			ReadTimeout     time.Duration `json:"readTimeout,omitempty"`
			WriteTimeout    time.Duration `json:"writeTimeout,omitempty"`
			ShutdownTimeout time.Duration `json:"shutdownTimeout,omitempty"`
		}
		DB struct {
			User       string `json:"user,omitempty"`
			Password   string `json:"password,omitempty"`
			Host       string `json:"host,omitempty"`
			Name       string `json:"name,omitempty"`
			DisableTLS bool   `json:"disableTLS,omitempty"`
		}
		Auth struct {
			KeyID string `json:"keyid,omitempty"`
			//			PrivateKeyFile string `conf:"default:/app-library/private.pem"`
			PrivateKeyFile string `json:"privateKeyFile,omitempty"`
			Algorithm      string `json:"algorithm,omitempty"`
		}
		OAuth struct {
			ClientID     string   `json:"clientID,omitempty"`
			ClientSecret string   `json:"clientSecret,omitempty"`
			Endpoint     string   `json:"endpoint,omitempty"`
			RedirectUrl  string   `json:"redirectUrl,omitempty"`
			Scopes       []string `json:"scopes,omitempty"`
			Issuer       string   `json:"issuer,omitempty"`
		}
		Zipkin struct {
			LocalEndpoint string  `json:"localEndpoint,omitempty"`
			ReporterURI   string  `json:"reporterURI,omitempty"`
			ServiceName   string  `json:"serviceName,omitempty"`
			Probability   float64 `json:"probability,omitempty"`
		}
	}

	provider := oidc.InsecureIssuerURLContext(ctx, cfg.OAuth.Issuer)
	log.Printf("main : provider context version %q", provider)
	// =========================================================================
	// App Starting

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(build)
	defer log.Println("main : Completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main : Config :\n%v\n", out)

	// =========================================================================
	// Initialize authentication support

	log.Println("main : Started : Initializing authentication support")

	authenticator, err := auth.OAuthenticate(cfg.OAuth.ClientID, cfg.OAuth.ClientSecret, cfg.OAuth.Endpoint, cfg.OAuth.RedirectUrl, cfg.OAuth.Scopes)
	if err != nil {
		return errors.Wrap(err, "constructing authenticator")
	}

	// =========================================================================
	// Start Database

	log.Println("main : Started : Initializing database support")

	db, err := database.Open(database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	})
	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}

	defer func() {
		log.Printf("main : Database Stopping : %s", cfg.DB.Host)
		db.Close()
	}()

	// =========================================================================
	// Start Tracing Support

	log.Println("main : Started : Initializing zipkin tracing support")

	localEndpoint, err := openzipkin.NewEndpoint(cfg.Zipkin.ServiceName, cfg.Zipkin.LocalEndpoint)
	if err != nil {
		return err
	}

	reporter := zipkinHTTP.NewReporter(cfg.Zipkin.ReporterURI)
	ze := zipkin.NewExporter(reporter, localEndpoint)

	trace.RegisterExporter(ze)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.ProbabilitySampler(cfg.Zipkin.Probability),
	})

	defer func() {
		log.Printf("main : Tracing Stopping : %s", cfg.Zipkin.LocalEndpoint)
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
		log.Printf("main : Debug Listening %s", cfg.Web.DebugHost)
		log.Printf("main : Debug Listener closed : %v", http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux))
	}()

	// =========================================================================
	// Start API Service

	log.Println("main : Started : Initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      handlers.API(build, shutdown, log, db, authenticator),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
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
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main : Graceful shutdown did not complete in %v : %v", cfg.Web.ShutdownTimeout, err)
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
