package category_test

import (
	category "github.com/book-library/internal/book-category"
	"github.com/book-library/internal/platform/auth"
	"github.com/book-library/internal/tests"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"testing"
	"time"
)

func TestCategory(t *testing.T)  {
	db, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to work with Book-category records.")
	{
		t.Log("\tWhen handling categories.")
		{
			ctx := tests.Context()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			// claims is information about the person making the request.
			claims := auth.NewClaims(
				"718ffbea-f4a1-4667-8ae3-b349da52675e", // This is just some random UUID.
				[]string{auth.RoleAdmin},
				now, time.Hour,
			)

			newcat := category.NewBookCategory{
				CategoryName:     "computer-science",
				NumberOfBooksIn:  1,
				NumberOfBooksOut: 2,
				DateCreated: now,
			}

			//test category creation
			cat, err := category.Create(ctx, now, newcat, claims, db)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create new book-category : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create new book-category.", tests.Success)

			//tests category retrieve
			savedCat, err := category.Retrieve(ctx, db, cat.ID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retreive book-category : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retreive a book-category by ID.", tests.Success)

			//tests if the save category is rendered
			if diff := cmp.Diff(savedCat, cat); diff != "" {
				t.Fatalf("\t%s\tShould get back the same book. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same book.", tests.Success)

			//tests category updated
			uctg := category.UpdateBookCategory{
				CategoryName: tests.StringPointer("computer-science"),
				NumberOfBooksIn:  tests.IntPointer(0),
				NumberOfBooksOut: tests.IntPointer(3),
				DateUpdated: tests.DatePointer(now),
			}

			if err := category.Update(ctx, savedCat.ID, uctg, now,  claims, db); err != nil{
				t.Fatalf("\t%s\tShould be able to update category : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould get back the updated category.", tests.Success)

			savedCat, err = category.Retrieve(ctx, db, cat.ID)
			if (err != nil) {
				t.Fatalf("\t%s\tShould be able to retreive updated category : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould get back the updated category.", tests.Success)

			if savedCat.CategoryName != *uctg.CategoryName {
				t.Errorf("\t%s\tShould be able to see updates to Name.", tests.Failed)
				t.Log("\t\tGot:", savedCat.CategoryName)
				t.Log("\t\tExp:", *uctg.CategoryName)
			} else {
				t.Logf("\t%s\tShould be able to see updates to categoryName.", tests.Success)
			}

			//test delete category
			if err := category.Delete(ctx, cat.ID, claims, db); err != nil {
				t.Fatalf("\t%s\tShould be able to delete category : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete category.", tests.Success)

			//check if category is retreivable
			savedCat, err = category.Retrieve(ctx, db, cat.ID)
			if errors.Cause(err) != category.ErrNotFound {
				t.Fatalf("\t%s\tShould be able NOT to retreive category : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to delete category.", tests.Success)
		}
	}
}