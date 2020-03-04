package mid

import (
	"context"
	"expvar"
	"net/http"
	"runtime"

	"github.com/book-library/internal/platform/web"
	"go.opencensus.io/trace"
)

//m is the global program counters for the application
//expvar provides a standardized interface to public variables, such as operation
//counters in servers. It exposes these variables via HTTP at /debug/vars in JSON format
var m = struct {
	gr  *expvar.Int
	req *expvar.Int
	err *expvar.Int
}{
	gr:  expvar.NewInt("goroutine"),
	req: expvar.NewInt("requests"),
	err: expvar.NewInt("errors"),
}

//Metrics retreive infos from http request and responses to update program counters.
func Metrics() web.Middleware {

	//actual middleware to be executed
	f := func(before web.Handler) web.Handler {

		//wrapped handler around the next one
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, param map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Metrics")
			defer span.End()

			err := before(ctx, w, r, param)

			//Increment the counter
			m.req.Add(1)

			//Increment the count for the number of the active goroutines every 100 requests.
			if m.req.Value()%100 == 0 {
				m.gr.Set(int64(runtime.NumGoroutine()))
			}

			//Increment the error counter
			if err != nil {
				m.err.Add(1)
			}

			////Return the error to be handled further up the chain
			return err
		}
		return h
	}
	return f
}
