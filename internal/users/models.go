package users

import (
	"time"

	"github.com/lib/pq"
)

// User represents someone with access to our system.
type User struct {
	ID           string         `db:"user_id" json:"id"`
	Name         string         `db:"name" json:"name"`
	Email        string         `db:"email" json:"email"`
	Roles        pq.StringArray `db:"roles" json:"roles"`
	PasswordHash []byte         `db:"password_hash" json:"-"`
	DateCreated  time.Time      `db:"date_created" json:"date_created"`
	DateUpdated  time.Time      `db:"date_updated" json:"date_updated"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required"`
	Roles           []string `json:"roles" validate:"required"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"password_confirm" validate:"eqfield=Password"`
}

// UpdateUser defines what information may be provided to modify an existing
// User. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
type UpdateUser struct {
	Name            *string  `json:"name"`
	Email           *string  `json:"email"`
	Roles           []string `json:"roles"`
	Password        *string  `json:"password"`
	PasswordConfirm *string  `json:"password_confirm" validate:"omitempty,eqfield=Password"`
}

// Session contains information needed to manage session-cookie/token.
type Session struct {
	Id     string    `db:"user_id" json:"user_id"`
	Token  string    `db:"token" json:"token"`
	Data   []byte    `db:"data" json:"data"`
	Expiry time.Time `db:"expiry" json:"expiry"`
}

// UpdateSession update the session-cookie/token.
type UpdateSession struct {
	Token   *string   `db:"token" json:"token"`
	NewData []byte    `db:"data" json:"data"`
	Expiry  time.Time `db:"expiry" json:"expiry"`
}
