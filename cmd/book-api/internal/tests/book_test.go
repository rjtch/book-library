package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/book-library/cmd/book-api/internal/handlers"
	"github.com/book-library/internal/platform/web"
	"github.com/book-library/internal/tests"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestProducts runs a series of tests to exercise Product behavior from the
// API level. The subtests all share the same database and application for
// speed and convenience. The downside is the order the tests are ran matters
// and one test may break if other tests are not ran before it. If a particular
// subtest needs a fresh instance of the application it can make it or it
// should be its own Test* function.
func TestBooks(t *testing.T) {
	test := tests.NewIntegration(t)
	defer test.Teardown()

	shutdown := make(chan os.Signal, 1)
	tests := BookTests{
		app:       handlers.API("develop", shutdown, test.Log, test.DB, test.Authenticator),
		userToken: test.Token("admin@example.com", "gophers"),
	}

	t.Run("postBook400", tests.postBook400)
	t.Run("getBook404", tests.getBook404)
	t.Run("postBook401", tests.postBook401)
}

// BookTests holds methods for each book subtest. This type allows
// passing dependencies for tests while still providing a convenient syntax
// when subtests are registered.
type BookTests struct {
	app       http.Handler
	userToken string
}

func (bt *BookTests) postBook400(t *testing.T) {
	r := httptest.NewRequest("POST", "/v1/books/create", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+bt.userToken)

	bt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new book can't be created with an invalid document.")
	{
		t.Log("\tTest 0:\tWhen using an incomplete book value.")
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tShould receive a status code of 400 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)

			// Inspect the response.
			var got web.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response to an error type : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to unmarshal the response to an error type.", tests.Success)

			// Define what we want to see.
			want := web.ErrorResponse{
				Error: "field validation error",
				Fields: []web.FieldError{
					{Field: "quantity", Error: "quantity is must be 1 or greater"},
				},
			}

			// We can't rely on the order of the field errors so they have to be
			// sorted. Tell the cmp package how to sort them.
			sorter := cmpopts.SortSlices(func(a, b web.FieldError) bool {
				return a.Field < b.Field
			})

			if diff := cmp.Diff(want.Error, got.Error, sorter); diff != "" {
				t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

func (bt *BookTests) getBook404(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/books/15b1d574-173a-4956-87a7-5c1796a10613", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+bt.userToken)

	bt.app.ServeHTTP(w, r)

	t.Log("Given the need to retrieve a book but can't be retrieved with an invalid ID.")
	{
		t.Log("\tTest 0:\tWhen using an invalid book-ID.")
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tShould receive a status code of 404 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 404 for the response.", tests.Success)

			// Inspect the response.
			var got web.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response to an error type : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to unmarshal the response to an error type.", tests.Success)

			// Define what we want to see.
			want := web.ErrorResponse{
				Error: "Book not found",
				Fields: []web.FieldError{
					{Field: "title", Error: "field validation error"},
				},
			}

			// We can't rely on the order of the field errors so they have to be
			// sorted. Tell the cmp package how to sort them.
			sorter := cmpopts.SortSlices(func(a, b web.FieldError) bool {
				return a.Field < b.Field
			})

			if diff := cmp.Diff(want.Error, got.Error, sorter); diff != "" {
				t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

func (bt *BookTests) postBook401(t *testing.T) {
	r := httptest.NewRequest("POST", "/v1/books/create", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	bt.userToken = ""

	r.Header.Set("Authorization", "Bearer "+bt.userToken)

	bt.app.ServeHTTP(w, r)

	t.Log("Given the need to retrieve a book but can't be retrieved with an invalid ID.")
	{
		t.Log("\tTest 0:\tWhen using an invalid book-ID.")
		{
			//TODO should and not 500 return 401 or 403
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("\t%s\tShould receive a status code of 404 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 404 for the response.", tests.Success)

			// Inspect the response.
			var got web.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response to an error type : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to unmarshal the response to an error type.", tests.Success)

			// Define what we want to see.
			want := web.ErrorResponse{
				Error: "Internal Server Error",
				Fields: []web.FieldError{
					{Field: "title", Error: "title is a required field"},
					{Field: "isbn", Error: "isbn is a required field"},
					{Field: "category", Error: "category is a required field"},
					{Field: "authors", Error: "authors is a required field"},
					{Field: "quantity", Error: "quantity is a required field"},
				},
			}

			// We can't rely on the order of the field errors so they have to be
			// sorted. Tell the cmp package how to sort them.
			sorter := cmpopts.SortSlices(func(a, b web.FieldError) bool {
				return a.Field < b.Field
			})

			if diff := cmp.Diff(want.Error, got.Error, sorter); diff != "" {
				t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}
