package tests

import (
	"encoding/json"
	"github.com/book-library/cmd/book-api/internal/handlers"
	"github.com/book-library/internal/platform/web"
	"github.com/book-library/internal/tests"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
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
	t.Run("postBook401", tests.postBook401)
	t.Run("getBook404", tests.getBook404)
	t.Run("getBook400", tests.getBook400)
	t.Run("deleteBookNotFound", tests.deleteBookNotFound)
	t.Run("putBook404", tests.putBook404)
	t.Run("crudBook", tests.crudBook)
}

// BookTests holds methods for each book subtest. This type allows
// passing dependencies for tests while still providing a convenient syntax
// when subtests are registered.
type BookTests struct {
	app       http.Handler
	userToken string
}

func (bt *BookTests) postBook400(t *testing.T) {
	r := httptest.NewRequest("POST", "/v1/books", strings.NewReader(`{}`))
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

			if diff := cmp.Diff(want, got, sorter); diff != "" {
				t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

func (bt *BookTests) postBook401(t *testing.T) {

}

func (bt *BookTests) getBook404(t *testing.T) {

}

func (bt *BookTests) getBook400(t *testing.T) {

}

func (bt *BookTests) deleteBookNotFound(t *testing.T) {

}

func (bt *BookTests) putBook404(t *testing.T) {

}

func (bt *BookTests) crudBook(t *testing.T) {

}