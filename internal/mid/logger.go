package mid

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/book-library/internal/platform/web"
	errors "github.com/pkg/errors"
	"go.opencensus.io/trace"
)

//Logger logs all coming request with specific informations
//format : TraceID : (code) VERB /path -> IP ADDR (latency)
func Logger(log *log.Logger) web.Middleware {

	//actual middleware to be executed
	f := func(before web.Handler) web.Handler {

		//wrapped handler around the next one
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Logger")
			defer span.End()

			// If the context is missing this value
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return errors.New("claims missing from context: HasRole called without/before Authenticate")
			}

			err := before(ctx, w, r, params)

			log.Printf("%s : (%d) : %s %s -> %s (%s)",
				v.TraceID,
				v.StatusCode,
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
				time.Since(v.Now),
			)

			//in case the error has been handle
			return err
		}
		return h
	}
	return f
}
