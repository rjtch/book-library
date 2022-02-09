package web

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

const (
	AllowOriginKey      string = "Access-Control-Allow-Origin"
	AllowCredentialsKey        = "Access-Control-Allow-Credentials"
	AllowHeadersKey            = "Access-Control-Allow-Headers"
	AllowMethodsKey            = "Access-Control-Allow-Methods"
	MaxAgeKey                  = "Access-Control-Max-Age"

	OriginKey         = "Origin"
	RequestMethodKey  = "Access-Control-Request-Method"
	RequestHeadersKey = "Access-Control-Request-Headers"
	ExposeHeadersKey  = "Access-Control-Expose-Headers"
)

//Respond converts Go value to JSON and sends it to the client
func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {

	// Set the status code for the request logger middleware.
	// If the context is missing this value, request the service
	// to be shutdown gracefully.
	v, ok := ctx.Value(KeyValues).(*Values)
	if !ok {
		return NewShutdownError("web value missing form context")
	}

	v.StatusCode = statusCode

	// If there is nothing to marshal then set status code and return.
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	// Convert the response value to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// Convert the response value to String
	token := reflect.ValueOf(data)

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")
	enableCors(&w)

	//write the status code to the response
	w.WriteHeader(statusCode)

	// Finally, we set the client cookie for "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	http.SetCookie(w, &http.Cookie{
		Name:     "Access-Token",
		Value:    token.String(),
		Expires:  time.Now().Add(30).UTC(),
		MaxAge:   600000000,
		Secure:   false,
		HttpOnly: true,
	})

	//Send the result back to the client
	if _, err := w.Write(jsonData); err != nil {
		return nil
	}

	return nil
}

//ResponseError sends errorful response back to the client
func ResponseError(ctx context.Context, w http.ResponseWriter, err error) error {

	// If the error was of the type *Error, the handler has
	// a specific status code and error to return.
	if webErr, ok := errors.Cause(err).(*Error); ok {
		er := ErrorResponse{
			Error:  webErr.Err.Error(),
			Fields: webErr.Fields,
		}

		if err := Respond(ctx, w, er, webErr.Status); err != nil {
			return err
		}
		return nil
	}

	// If not, the handler sent any arbitrary error value so use 500.
	er := ErrorResponse{
		Error: http.StatusText(http.StatusInternalServerError),
	}
	if err := Respond(ctx, w, er, http.StatusInternalServerError); err != nil {
		return err
	}
	return nil
}

//enableCors enables cross origin control
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set(AllowOriginKey, "*")
	(*w).Header().Set(AllowCredentialsKey, "*")
	(*w).Header().Set(AllowHeadersKey, "*")
	(*w).Header().Set(OriginKey, "*")
}
