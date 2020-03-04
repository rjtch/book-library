package mid

import (
	"context"
	"log"
	"net/http"

	"github.com/book-library/internal/platform/web"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
)

//Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
func Errors(log *log.Logger) web.Middleware {

	//actual middleware to be execute
	f := func(before web.Handler) web.Handler {

		//wrapped handler around the next one
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Errors")
			defer span.End()

			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return errors.New("claims missing from context: HasRole called without/before Authenticate")
			}

			if err := before(ctx, w, r, params); err != nil {

				//log the error
				log.Printf("%s : ERROR : %+v ", v.TraceID, err)

				//respond to the error
				if err := web.ResponseError(ctx, w, err); err != nil {
					return err
				}

				//In case we receive a shutdown signal
				if ok := web.IsShutdown(err); ok {
					return err
				}
			}

			//in case the error has been handle
			return nil
		}
		return h
	}
	return f
}
