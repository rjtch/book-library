package users_test

import (
	"testing"
	"time"

	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/tests"
	"github.com/book-library/internal/users"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

// TestUser validates the full set of CRUD operations on User values.
func TestUnitUser(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to work with User records.")
	{
		t.Log("\tWhen handling a single User.")
		{
			ctx := tests.Context()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)
			nu := users.NewUser{
				Name:            "Bill Kennedy",
				Email:           "bill@ardanlabs.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "gophers",
				PasswordConfirm: "gophers",
			}

			//tests user creation
			u, err := users.Create(ctx, db, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			savedU, err := users.Retrieve(ctx, db, u.ID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user by ID: %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user by ID.", tests.Success)

			if diff := cmp.Diff(u, savedU); diff != "" {
				t.Fatalf("\t%s\tShould get back the same user. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same user.", tests.Success)

			upd := users.UpdateUser{
				Name:  tests.StringPointer("Jacob Walker"),
				Email: tests.StringPointer("jacob@ardanlabs.com"),
			}

			if err := users.Update(ctx, db, u.ID, upd, now); err != nil {
				t.Fatalf("\t%s\tShould be able to update user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to update user.", tests.Success)

			savedU, err = users.Retrieve(ctx, db, u.ID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user.", tests.Success)

			if savedU.Name != *upd.Name {
				t.Errorf("\t%s\tShould be able to see updates to Name.", tests.Failed)
				t.Log("\t\tGot:", savedU.Name)
				t.Log("\t\tExp:", *upd.Name)
			} else {
				t.Logf("\t%s\tShould be able to see updates to Name.", tests.Success)
			}

			if savedU.Email != *upd.Email {
				t.Errorf("\t%s\tShould be able to see updates to Email.", tests.Failed)
				t.Log("\t\tGot:", savedU.Email)
				t.Log("\t\tExp:", *upd.Email)
			} else {
				t.Logf("\t%s\tShould be able to see updates to Email.", tests.Success)
			}

			if err := users.Delete(ctx, db, u.ID); err != nil {
				t.Fatalf("\t%s\tShould be able to delete user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete user.", tests.Success)

			savedU, err = users.Retrieve(ctx, db, u.ID)
			if errors.Cause(err) != users.ErrNotFound {
				t.Fatalf("\t%s\tShould NOT be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to retrieve user.", tests.Success)
		}
	}
}

// TestAuthenticate validates the behavior around authenticating users.
func TestAuthenticate(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to authenticate users")
	{
		t.Log("\tWhen handling a single User.")
		{
			ctx := tests.Context()

			nu := users.NewUser{
				Name:            "Anna Walker",
				Email:           "anna@ardanlabs.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "goroutines",
				PasswordConfirm: "goroutines",
			}

			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			_, err := users.Create(ctx, db, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			claims, err := users.Authenticate(ctx, db, now, "anna@ardanlabs.com", "goroutines")
			if err != nil {
				t.Fatalf("\t%s\tShould be able to generate claims : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to generate claims.", tests.Success)

			want := auth.NewClaims(jwt.Token{}, []string{auth.RoleAdmin, auth.RoleUser}, "718ffbea-f4a1-4667-8ae3-b349da52675e")
			if diff := cmp.Diff(want, claims); diff != "" {
				t.Fatalf("\t%s\tShould get back the expected claims. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the expected claims.", tests.Success)
		}
	}
}
