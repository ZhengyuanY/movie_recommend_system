package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

// GetMovieById
// -- expect valid uuid in db return matched docs
func TestValidUUIDInDBSuccess(t *testing.T) {
	testTime := time.Now()
	timeHelper := func() time.Time {
		return testTime
	}
	testDbAddr := "http://mockES"
	testMovieIndex := "movies"
	testDocID := "test_id"
	testRequestURL := fmt.Sprintf("%s/%s/_doc/%s?pretty", testDbAddr, testMovieIndex, testDocID)
	testDocBody := `[{"_id": test_id, "name": "test_id"}]`
	acMux := &autocompleteMux{
		ServeMux:   http.NewServeMux(),
		dbAddr:     testDbAddr,
		movieIndex: testMovieIndex,
		timeModule: timeHelper,
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Exact URL match
	httpmock.RegisterResponder("GET", testRequestURL, httpmock.NewStringResponder(200, testDocBody))

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", testRequestURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	acMux.GetMovieByIDHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler function returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if rr.Body.String() != testDocBody+fmt.Sprintf("%v", testTime) {
		t.Errorf("Returned wrong doc:\ngot %v\nwant %v", rr.Body.String(), testDocBody)
	}
}

// -- expect valid uuid in db but db failure return 500
func TestValidUUIDInDBWithDBFailure(t *testing.T) {
	testDbAddr := "http://mockES"
	testMovieIndex := "movies"
	testDocID := "test_id"
	testRequestURL := fmt.Sprintf("%s/%s/_doc/%s?pretty", testDbAddr, testMovieIndex, testDocID)
	acMux := &autocompleteMux{
		ServeMux:   http.NewServeMux(),
		dbAddr:     testDbAddr,
		movieIndex: testMovieIndex,
		timeModule: time.Now,
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Simulate a database failure
	httpmock.RegisterResponder("GET", testRequestURL, httpmock.NewStringResponder(500, ""))

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", testRequestURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	acMux.GetMovieByIDHandler(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler function returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

// -- expect valid uuid in db but network failure return 500
func TestValidUUIDInDBWithNetworkFailure(t *testing.T) {
	testDbAddr := "http://mockES"
	testMovieIndex := "movies"
	testDocID := "test_id"
	testRequestURL := fmt.Sprintf("%s/%s/_doc/%s?pretty", testDbAddr, testMovieIndex, testDocID)
	acMux := &autocompleteMux{
		ServeMux:   http.NewServeMux(),
		dbAddr:     testDbAddr,
		movieIndex: testMovieIndex,
		timeModule: time.Now,
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Simulate a network failure
	httpmock.RegisterResponder("GET", testRequestURL, httpmock.NewErrorResponder(fmt.Errorf("network failure")))

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", testRequestURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	acMux.GetMovieByIDHandler(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler function returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

// -- expect valid uuid out of db return 404
func TestValidUUIDOutOfDBReturns404(t *testing.T) {
	testDbAddr := "http://mockES"
	testMovieIndex := "movies"
	testDocID := "test_id_not_in_db"
	testRequestURL := fmt.Sprintf("%s/%s/_doc/%s?pretty", testDbAddr, testMovieIndex, testDocID)
	acMux := &autocompleteMux{
		ServeMux:   http.NewServeMux(),
		dbAddr:     testDbAddr,
		movieIndex: testMovieIndex,
		timeModule: time.Now,
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Simulate a 404 Not Found response from the database
	httpmock.RegisterResponder("GET", testRequestURL, httpmock.NewStringResponder(http.StatusNotFound, ""))

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", testRequestURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	acMux.GetMovieByIDHandler(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler function returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

// -- expect valid uuid out of db but db failure return 500
func TestValidUUIDOutOfDBWithDBFailure(t *testing.T) {
	testDbAddr := "http://mockES"
	testMovieIndex := "movies"
	testDocID := "test_id_not_in_db"
	testRequestURL := fmt.Sprintf("%s/%s/_doc/%s?pretty", testDbAddr, testMovieIndex, testDocID)
	acMux := &autocompleteMux{
		ServeMux:   http.NewServeMux(),
		dbAddr:     testDbAddr,
		movieIndex: testMovieIndex,
		timeModule: time.Now,
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Simulate a database failure when the document is not found
	httpmock.RegisterResponder("GET", testRequestURL, httpmock.NewStringResponder(500, ""))

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", testRequestURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	acMux.GetMovieByIDHandler(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler function returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

// -- return invalid uuid error on invalid uuid input
func TestInvalidUUIDInput(t *testing.T) {
	testDbAddr := "http://mockES"
	testMovieIndex := "movies"
	invalidDocID := strings.Repeat("a", 513) // Invalid ID longer than 512 bytes
	testRequestURL := fmt.Sprintf("%s/%s/_doc/%s?pretty", testDbAddr, testMovieIndex, invalidDocID)
	acMux := &autocompleteMux{
		ServeMux:   http.NewServeMux(),
		dbAddr:     testDbAddr,
		movieIndex: testMovieIndex,
		timeModule: time.Now,
	}

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", testRequestURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	acMux.GetMovieByIDHandler(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler function returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

// -- expect extra path segments return 500
func TestExtraPathSegments(t *testing.T) {
	testDbAddr := "http://mockES"
	testMovieIndex := "movies"
	testDocID := "test_id"
	testRequestURL := fmt.Sprintf("%s/%s/_doc/%s/extra_segment", testDbAddr, testMovieIndex, testDocID)
	acMux := &autocompleteMux{
		ServeMux:   http.NewServeMux(),
		dbAddr:     testDbAddr,
		movieIndex: testMovieIndex,
		timeModule: time.Now,
	}

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", testRequestURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	acMux.GetMovieByIDHandler(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler function returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

// -- expect valid uuid in db with query parameters return matched docs ignoring query parameters
func TestValidUUIDInDBWithQueryParams(t *testing.T) {
	testTime := time.Now()
	timeHelper := func() time.Time {
		return testTime
	}
	testDbAddr := "http://mockES"
	testMovieIndex := "movies"
	testDocID := "test_id"
	testRequestURL := fmt.Sprintf("%s/%s/_doc/%s?pretty", testDbAddr, testMovieIndex, testDocID)
	testDocBody := `[{"_id": test_id, "name": "test_id"}]`
	acMux := &autocompleteMux{
		ServeMux:   http.NewServeMux(),
		dbAddr:     testDbAddr,
		movieIndex: testMovieIndex,
		timeModule: timeHelper,
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Exact URL match ignoring query parameters
	httpmock.RegisterResponder("GET", testRequestURL, httpmock.NewStringResponder(200, testDocBody))

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", testRequestURL+"&extra_param=value", nil)
	if err != nil {
		t.Fatal(err)
	}
	acMux.GetMovieByIDHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler function returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Body.String() != testDocBody+fmt.Sprintf("%v", testTime) {
		t.Errorf("Returned wrong doc:\ngot %v\nwant %v", rr.Body.String(), testDocBody)
	}
}

// Autocomplete
// -- expect query as a whole phrase return all the docs with exact match in any field
// -- expect query with leading spaces return all the docs with exact match of query without leading spaces in any field
// -- expect query with trailing spaces return all the docs with exact match of query without trailing spaces in any field
// -- expect query with multiple consecutive spaces in between words return all the docs with exact match of query with single spaces in between words in any field
// -- expect query as a phrase combining multiple words not return any doc with shuffled combination of words
// -- expect query as a whole phrase return empty when no exact match in any field
// -- expect query with non-ASCII characters return all the docs with exact match of the query including non-ASCII characters in any field
// -- expect query with wildcards return all the docs that match the wildcard pattern in any field
// -- expect empty query return an error or appropriate message indicating the query is required
// -- expect query with a very long string (over typical limits) return an error or handle appropriately
// -- expect query containing SQL-like injection attempts return an error or handle safely without executing any injected commands
