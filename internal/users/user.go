package users

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/book-library/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("User not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrAuthenticationFailure occurs when a users attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("Authentication failed")

	// ErrForbidden occurs when a users tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("Attempted action is not allowed")
)

const (
	defaultJWTCookieName  = "SESSION-COOKIE"
	defaultXSRFCookieName = "XSRF-TOKEN"
)

// List retrieves a list of existing users from the database.
func List(ctx context.Context, claims auth.Claims, db *sqlx.DB) ([]User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.users.List")
	defer span.End()

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	if !claims.HasRole(auth.RoleAdmin) {
		return nil, ErrForbidden
	}

	users := []User{}
	const q = `SELECT * FROM users`

	if err := db.SelectContext(ctx, &users, q); err != nil {
		return nil, errors.Wrap(err, "selecting users")
	}

	return users, nil
}

// Retrieve gets the specified users from the database.
func Retrieve(ctx context.Context, claims auth.Claims, db *sqlx.DB, id string) (*User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.users.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	if !claims.HasRole(auth.RoleAdmin) && claims.Subject != id {
		return nil, ErrForbidden
	}

	var u User
	const q = `SELECT * FROM users WHERE user_id = $1`
	if err := db.GetContext(ctx, &u, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting users %q", id)
	}

	return &u, nil
}

// Retrieve gets the actual user from the database.
func RetrieveMe(ctx context.Context, claims auth.Claims, db *sqlx.DB)(*User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.users.RetrieveMe")
	defer span.End()

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return nil, errors.Wrap(nil, "claims missing from context")
	}

	var u User
	const q = `SELECT current_user`
	if err := db.GetContext(ctx, &u, q); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrapf(err, "selecting users %q")
	}

	return &u, nil
}

// Create inserts a new users into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewUser, now time.Time) (*User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.users.Create")
	defer span.End()

	hash, err := bcrypt.GenerateFromPassword([]byte(n.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "generating password hash")
	}

	u := User{
		ID:           uuid.New().String(),
		Name:         n.Name,
		Email:        n.Email,
		PasswordHash: hash,
		Roles:        n.Roles,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `INSERT INTO users
		(user_id, name, email, password_hash, roles, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = db.ExecContext(
		ctx, q,
		u.ID, u.Name, u.Email,
		u.PasswordHash, u.Roles,
		u.DateCreated, u.DateUpdated,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting users")
	}

	return &u, nil
}

// Update replaces a users document in the database.
func Update(ctx context.Context, claims auth.Claims, db *sqlx.DB, id string, upd UpdateUser, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.users.Update")
	defer span.End()

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	if !claims.HasRole(auth.RoleUser) && claims.Subject != id {
		return ErrForbidden
	}

	u, err := Retrieve(ctx, claims, db, id)
	if err != nil {
		return err
	}

	if upd.Name != nil {
		u.Name = *upd.Name
	}
	if upd.Email != nil {
		u.Email = *upd.Email
	}
	if upd.Roles != nil {
		u.Roles = upd.Roles
	}
	if upd.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*upd.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		u.PasswordHash = pw
	}

	u.DateUpdated = now

	const q = `UPDATE users SET
		"name" = $2,
		"email" = $3,
		"roles" = $4,
		"password_hash" = $5,
		"date_updated" = $6
		WHERE user_id = $1`
	_, err = db.ExecContext(ctx, q, id,
		u.Name, u.Email, u.Roles,
		u.PasswordHash, u.DateUpdated,
	)
	if err != nil {
		return errors.Wrap(err, "updating users")
	}

	return nil
}

// Delete removes a users from the database.
func Delete(ctx context.Context, db *sqlx.DB, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.users.Delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM users WHERE user_id = $1`

	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting users %s", id)
	}

	return nil
}

// Authenticate finds a users by their email and verifies their password. On
// success it returns a Claims value representing this users. The claims can be
// used to generate a token for future authentication.
func Authenticate(ctx context.Context, db *sqlx.DB, now time.Time, email, password string) (auth.Claims, error) {
	ctx, span := trace.StartSpan(ctx, "internal.users.Authenticate")
	defer span.End()

	const q = `SELECT * FROM users WHERE email = $1`
	var u User
	if err := db.GetContext(ctx, &u, q, email); err != nil {
		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated users which emails are in the system.
		if err == sql.ErrNoRows {
			return auth.Claims{}, ErrAuthenticationFailure
		}

		return auth.Claims{}, errors.Wrap(err, "selecting single users")
	}

	// Compare the provided password with the saved hash. Use the bcrypt
	// comparison function so it is cryptographically secure.
	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		return auth.Claims{}, ErrAuthenticationFailure
	}

	// If we are this far the request is valid. Create some claims for the users
	// and generate their token.
	claims := auth.NewClaims(u.ID, u.Roles, now, time.Hour)

	//save the claim as session-token or drop if claim is nil
	const t = `INSERT INTO sessions (user_id, token, data, expiry) VALUES ($1, $2, $3, $4)`
	//convert claims in string
	str := fmt.Sprint(claims)
	timer := time.Now().Add(10)
	//encode converted string in base64
	encoded := base64.StdEncoding.EncodeToString([]byte(str))
	_, err := db.ExecContext(
		ctx, t, u.ID, defaultJWTCookieName, []byte(encoded), timer)
	if err != nil {
		return auth.Claims{}, errors.Wrap(err, "Session expired or not existed")
	}

	return claims, nil
}

func RefreshesToken(ctx context.Context, db *sqlx.DB, user_id string) (auth.Claims, error) {
	ctx, span := trace.StartSpan(ctx, "internal.users.RefreshToken")
	defer span.End()

	const q = `SELECT * FROM sessions WHERE user_id = $1`
	var tk Session
	if err := db.GetContext(ctx, &tk, q, user_id); err != nil {
		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated users which emails are in the system.
		if err == sql.ErrNoRows {
			return auth.Claims{}, ErrNotFound
		}

		return auth.Claims{}, errors.Wrap(err, "selecting single token")
	}

	var claim auth.Claims
	err := json.Unmarshal(tk.Data, &claim)
	if err != nil {
		return auth.Claims{}, errors.Wrap(err, "unable to convert byte data back")
	}
	if IsExpired(claim) {
		tk.Expiry = time.Now().Add(3600)
	}

	//save the claim as session-token or drop if claim is nil
	const t = `INSERT INTO sessions (token, data, expiry) VALUES ($1, $2, $3)`
	_, err = db.ExecContext(
		ctx, t, tk.Token, tk.Data, tk.Expiry)
	if err != nil {
		return auth.Claims{}, errors.Wrap(err, "Session expired or not existed")
	}

	return claim, nil
}

//Logout deletes user's session-token from the database which invalidates all existing cookies
//in browsers
func Logout(ctx context.Context, db *sqlx.DB, user_id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.users.Logout")
	defer span.End()

	//TODO make sure user is not logged in many times
	const q = `SELECT * FROM sessions WHERE user_id = $1`
	var tk Session
	if err := db.GetContext(ctx, &tk, q, user_id); err != nil {
		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated users which emails are in the system.
		if err == sql.ErrNoRows {
			return ErrNotFound
		}

		return errors.Wrap(err, "selecting single token")
	}

	//save the claim as session-token or drop if claim is nil
	const t = `DELETE FROM sessions WHERE user_id = $1`
	_, err := db.ExecContext(
		ctx, t, user_id)
	if err != nil {
		return errors.Wrap(err, "Deleting session-token")
	}

	return nil
}

//IsExpired verifies iif the given claim has expired or not.
func IsExpired(claims auth.Claims) bool {
	return !claims.VerifyExpiresAt(time.Now().Unix(), true)
}

func IsLoggedOut(ctx context.Context, db *sqlx.DB, user_id string) (bool, error){
	ctx, span := trace.StartSpan(ctx, "internal.users.IsLoggedOut")
	defer span.End()

	const q = `SELECT * FROM sessions WHERE user_id = $1`
	var tk Session
	if err := db.GetContext(ctx, &tk, q, user_id); err != nil {
		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated users which emails are in the system.
		if err == sql.ErrNoRows {
			return false, ErrNotFound
		}

		return false, errors.Wrap(err, "selecting single token")
	}

	if &tk == nil {
		return true, nil
	}

	return true, nil
}