package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/book-library/internal/mid"
	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/platform/web"
	"github.com/jmoiron/sqlx"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, db *sqlx.DB, authenticator *auth.Authenticator) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register health check endpoint. This route is not authenticated.
	check := Check{
		build: build,
		db:    db,
	}

	app.Handle("GET", "/v1/health", check.Health)

	// Register users management and authentication endpoints.
	u := User{
		Db:            db,
		authenticator: authenticator,
	}

	app.Handle("GET", "/v1/users/all", u.List, mid.Authentication(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("POST", "/v1/users/create", u.Create, mid.Authentication(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/users/:id", u.Retrieve, mid.Authentication(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("PUT", "/v1/users/:id/update", u.Update, mid.Authentication(authenticator), mid.HasRole(auth.RoleUser))
	app.Handle("DELETE", "/v1/users/:id/delete", u.Delete, mid.Authentication(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/users/:user-id/me", u.RetrieveMe, mid.Authentication(authenticator), mid.HasRole(auth.RoleUser))

	// This routes are not authenticated
	app.Handle("POST", "/v1/users/token", u.TokenAuthenticator)
	app.Handle("GET", "/v1/users/refresh-token", u.RefreshToken)
	app.Handle("POST", "/v1/users/:user_id/logout", u.Logout)

	// Register books endpoints.
	bk := Book{
		db: db,
	}
	app.Handle("GET", "/v1/books/all", bk.List, mid.Authentication(authenticator))
	app.Handle("GET", "/v1/books/title", bk.RetrieveByTitle, mid.Authentication(authenticator))
	app.Handle("POST", "/v1/books/create", bk.Create, mid.Authentication(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/books/:id", bk.Retrieve, mid.Authentication(authenticator))
	app.Handle("PUT", "/v1/books/:id/update", bk.Update, mid.Authentication(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("DELETE", "/v1/books/:id/delete", bk.Delete, mid.Authentication(authenticator), mid.HasRole(auth.RoleAdmin))

	// Register book-category endpoints.
	ct := BookCategory{
		db: db,
	}
	app.Handle("GET", "/v1/categories/all", ct.List, mid.Authentication(authenticator))
	app.Handle("POST", "/v1/categories/create", ct.Create, mid.Authentication(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("PUT", "/v1/categories/:id/update", ct.Update, mid.Authentication(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("DELETE", "/v1/categories/:id/delete", ct.Delete, mid.Authentication(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/categories/:id", ct.Retreive, mid.Authentication(authenticator), mid.HasRole(auth.RoleUser))

	// Register loans endpoints.
	l := Loan{
		db: db,
	}
	app.Handle("GET", "/v1/loans/:user_id/all", l.List, mid.Authentication(authenticator), mid.HasRole(auth.RoleUser))
	app.Handle("POST", "/v1/loans/:user_id/init", l.Create, mid.Authentication(authenticator), mid.HasRole(auth.RoleUser))
	app.Handle("PUT", "/v1/loans/:user_id/update/:id", l.Update, mid.Authentication(authenticator), mid.HasRole(auth.RoleUser))
	app.Handle("DELETE", "/v1/loans/:user_id/delete/:id", l.Delete, mid.Authentication(authenticator), mid.HasRole(auth.RoleUser))
	app.Handle("GET", "/v1/loans/:user_id/retrieve/:id", l.Retrieve, mid.Authentication(authenticator), mid.HasRole(auth.RoleUser))

	return app
}
