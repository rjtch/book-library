package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
)

//ctxKey represents the type of value for the context key.
type ctxKey int

// KeyValues is how request values are stored/retrieved.
const KeyValues ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

// A Handler is a type that handles an http request within the application
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error

//App is the entrypoint into our application and what configures our context object
//for each of our http.handlers.
type App struct {
	*httptreemux.TreeMux
	och      *ochttp.Handler
	shutdown chan os.Signal
	mw       []Middleware
}

//NewApp creates an App value that handle a set of routes for the application
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	app := App{
		TreeMux:  httptreemux.New(),
		shutdown: shutdown,
		mw:       mw,
	}

	// Create an OpenCensus HTTP Handler which wraps the router. This will start
	// the initial span and annotate it with information about the request/response.
	//for more infos https://w3c.github.io/trace-context/
	app.och = &ochttp.Handler{
		Handler: app.TreeMux,

		// Propagation defines how traces are propagated. If unspecified,
		// B3 propagation will be used.
		// HTTPFormat implementations propagate span contexts
		// in HTTP requests.
		Propagation: &tracecontext.HTTPFormat{},
	}

	return &app
}

// SignalShutdown is used to gracefully shutdown the app when an integrity
// issue is identified.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// Handle is our mechanism for mounting Handlers for a given HTTP verb and path
// pair, this makes for really easy, convenient routing.
func (a *App) Handle(verb, path string, handler Handler, mw ...Middleware) {

	//first wraphandler specific middleware around this handler
	handler = wrapMiddleware(mw, handler)

	//second wraphandler chains all middleware with application middlware
	handler = wrapMiddleware(a.mw, handler)

	// The function to execute for each request.
	h := func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		ctx, span := trace.StartSpan(r.Context(), "internal.platform.web")
		defer span.End()

		// Set the context with the required values to
		// process the request.
		v := Values{
			TraceID: span.SpanContext().TraceID.String(),
			Now:     time.Now(),
		}
		ctx = context.WithValue(ctx, KeyValues, &v)

		// Call the wrapped handler functions.
		err := handler(ctx, w, r, params)
		if err != nil {
			fmt.Println(err, "error while executing request")
			a.SignalShutdown()
			return
		}
	}

	// Add this handler for the specified verb and route.
	a.TreeMux.Handle(verb, path, h)

}

// ServeHTTP implements the http.Handler interface. It overrides the ServeHTTP
// of the embedded TreeMux by using the ochttp.Handler instead. That Handler
// wraps the TreeMux handler so the routes are served.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.och.ServeHTTP(w, r)
}
