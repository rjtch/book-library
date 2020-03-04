package mid

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"

	errors "github.com/pkg/errors"

	"github.com/book-library/internal/platform/web"
)

//Panics recovers from panics and converts the panics to an error so it is
//reported in Metrics and handled in Errors
func Panics(log *log.Logger) web.Middleware {

	//This is the actual middleware function to be executed
	f := func(after web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, param map[string]string) (err error) {

			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}

			//Defer a function to recover from a panic and set the err return
			//variable after fact.
			defer func() {
				if r := recover(); r != nil {
					err = errors.Errorf("panic: %v", r)

					//log the Go stack trace for this panic'd goroutine
					log.Printf("%s : \n%s", v.TraceID, debug.Stack())
				}
			}()

			//Call the next Handler and set its return value in the err variable.
			return after(ctx, w, r, param)
		}
		return h
	}
	return f
}
