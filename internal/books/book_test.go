package books_test

import (
	"testing"
	"time"

	"github.com/book-library/internal/books"
	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/tests"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

// TestBook validates the full set the CRUD operations on Book values
func TestBook(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to work with Book records.")
	{
		t.Log("\tWhen handling books.")
		{
			ctx := tests.Context()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			// claims is information about the person making the request.
			claims := auth.NewClaims(
				auth.RoleAdmin,
				[]string{auth.RoleAdmin, auth.RoleUser},
				now, time.Hour,
				"718ffbea-f4a1-4667-8ae3-b349da52675e", // This is just some random UUID.
			)

			nb := books.NewBook{
				Title:       "Go programming",
				ISBN:        "bcn22",
				Category:    "computer-science",
				Description: "Learn go the simplest way",
				Authors:     "Bill Kenedy",
				Quantity:    2,
			}

			//tests book creation
			bk, err := books.Create(ctx, now, nb, claims, db)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create book : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create book.", tests.Success)

			//tests book retrieve
			savedBk, err := books.Retrieve(ctx, bk.ID, db)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retreive book : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retreive a book by ID.", tests.Success)

			//tests if the save book is rendered
			if diff := cmp.Diff(bk, savedBk); diff != "" {
				t.Fatalf("\t%s\tShould get back the same book. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same book.", tests.Success)

			udbk := books.UpdateBook{
				Description: tests.StringPointer("Learn go the simplest way in 1 month"),
				Authors:     tests.StringPointer("Bill Kennedy"),
				Category:    tests.StringPointer("computer-science"),
				Quantity:    tests.IntPointer(3),
				DateUpdated: tests.DatePointer(now),
			}

			//test update book
			if err := books.Update(ctx, savedBk.ID, udbk, now, claims, db); err != nil {
				t.Fatalf("\t%s\tShould be able to update book : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould get back the updated book.", tests.Success)

			savedBk, err = books.Retrieve(ctx, bk.ID, db)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retreive book : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retreive a book by ID.", tests.Success)

			if savedBk.Authors != *udbk.Authors {
				t.Errorf("\t%s\tShould be able to see updates to Name.", tests.Failed)
				t.Log("\t\tGot:", savedBk.Authors)
				t.Log("\t\tExp:", *udbk.Authors)
			} else {
				t.Logf("\t%s\tShould be able to see updates to Authors.", tests.Success)
			}

			//test delete book
			if err := books.Delete(ctx, bk.ID, claims, db); err != nil {
				t.Fatalf("\t%s\tShould be able to delete book : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete book.", tests.Success)

			//test check if book is retrievable
			savedBk, err = books.Retrieve(ctx, bk.ID, db)
			if errors.Cause(err) != books.ErrNotFound {
				t.Fatalf("\t%s\tShould be able NOT to retreive book : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to delete book.", tests.Success)
		}
	}
}
