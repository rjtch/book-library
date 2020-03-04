package loans_test

import (
	"fmt"
	category "github.com/book-library/internal/book-category"
	"github.com/book-library/internal/books"
	loans "github.com/book-library/internal/loan"
	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/tests"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"testing"
	"time"
)

func TestLoan(t *testing.T) {

	db, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to work with Loan records.")
	{
		t.Log("\tWhen handling loans.")
		{
			ctx := tests.Context()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			// claims is information about the person making the request.
			claims := auth.NewClaims(
				"718ffbea-f4a1-4667-8ae3-b349da52675e", // This is just some random UUID.
				[]string{auth.RoleAdmin, auth.RoleUser},
				now, time.Hour,
			)

			newcat := category.NewBookCategory{
				CategoryName:     "computer-science",
				NumberOfBooksIn:  3,
				NumberOfBooksOut: 0,
				DateCreated: now,
			}

			//test category creation
			cat, err := category.Create(ctx, now, newcat, claims, db)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create new book-category : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create new book-category.", tests.Success)

			nb := books.NewBook{
				Title:       "Go programming",
				ISBN:        "bcn22",
				Category:    cat.CategoryName,
				Description: "Learn go the simplest way",
				Authors:     "Bill Kenedy",
				Quantity:    cat.NumberOfBooksIn,
			}

			//tests book creation
			bk, err := books.Create(ctx, now, nb, claims, db)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create book : %s.", tests.Failed, err)
			}

			t.Logf("\t%s\tShould be able to create book.", tests.Success)

			savedbk, err := books.Retrieve(ctx, bk.ID, db)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retreive book : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retreive book.", tests.Success)

			nl := loans.NewLoan{
				BookTitle:    savedbk.Title,
				BookISBN:     savedbk.ISBN,
				BookID:       savedbk.ID,
				BookQuantity: 1,
			}


			bks, errB := books.List(ctx, db)
			if errB != nil {
				t.Fatalf("\t%s\tShould be able to retreive loan : %s.", tests.Failed, err)
				fmt.Printf("%v", bks)
			}

			//test loan creation
			 ln, err := loans.InitNewLoan(ctx, claims, nl, now, nl.BookID, db)
			 if err != nil {
				t.Fatalf("\t%s\tShould be able to create new loan : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create new loan.", tests.Success)

			//test loan retrieve
			savedl, err := loans.Retrieve(ctx, claims, ln.ID, db)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retreive loan : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retreive loan.", tests.Success)

			//tests if the save loan is rendered
			if diff := cmp.Diff(ln.ID, savedl.ID); diff != "" {
				t.Fatalf("\t%s\tShould get back the same loan. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same loan.", tests.Success)

			ul := loans.UpdateLoan{
				BookISBN:     tests.StringPointer("bcn22"),
				ReturnDate:  tests.DatePointer(now.Add(30).UTC()),
				BookQuantity: tests.IntPointer(savedl.BookQuantity),
			}

			//test update loan
			if err := loans.Update(ctx, savedl.ID, ul, now, claims, db); err != nil {
				t.Fatalf("\t%s\tShould be able to update loan : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould get back the updated loan.", tests.Success)

			//test retrieve updated loan
			uln, err := loans.Retrieve(ctx, claims, savedl.ID, db)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retreive loan : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retreive a loan by ID.", tests.Success)

			if uln.BookISBN != *ul.BookISBN {
				t.Errorf("\t%s\tShould be able to see updates to isbn.", tests.Failed)
				t.Log("\t\tGot:", savedl.BookISBN)
				t.Log("\t\tExp:", *ul.BookISBN)
			} else {
				t.Logf("\t%s\tShould be able to see updates to isbn.", tests.Success)
			}

			//test delete loan
			if err := loans.EndUpALoan(ctx, claims, now ,uln.ID, db); err != nil {
				t.Fatalf("\t%s\tShould be able to delete loan : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete loan.", tests.Success)

			//test check if loan is retrievable
			savedl, err = loans.Retrieve(ctx, claims, uln.ID, db)
			if errors.Cause(err) != loans.ErrNotFound  {
				t.Fatalf("\t%s\tShould be able NOT to retreive loan : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to retreive loan.", tests.Success)
		}
	}
}
