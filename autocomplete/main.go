package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

type autocompleteMux struct {
	*http.ServeMux
	dbAddr     string
	movieIndex string
	timeModule func() time.Time
}

// statusErrorMap maps an internal error code to external error code
// -- only 404 maps to 404
// -- return 500 otherwise
func (acMux *autocompleteMux) statusErrorMap(statusCode int) int {
	if statusCode == http.StatusOK {
		slog.Warn("Not error code")
		return http.StatusInternalServerError
	}
	if statusCode == http.StatusNotFound {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

// GetMovieByIDHandler handles requests to the URL "/movies/id/$id"
// -- return movie json files on success
func (acMux *autocompleteMux) GetMovieByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		slog.Error("Incorrect number of  path segments")
		http.Error(w, "Incorrect number of  path segments", http.StatusBadRequest)
		return
	}
	id := parts[3]

	// Check if ID is less than 512 bytes
	if len(id) > 512 {
		slog.Error("Movie ID exceeds allowed length")
		http.Error(w, "Movie ID exceeds allowed length", http.StatusBadRequest)
		return
	}

	// Construct the URL for the GET request
	url := fmt.Sprintf("%s/%s/_doc/%s?pretty", acMux.dbAddr, acMux.movieIndex, id)

	// Send the Get request
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("Error sending request to database: %s", err)
		http.Error(w, fmt.Sprintf("Error getting movie ID=%s", id), acMux.statusErrorMap(http.StatusInternalServerError))
		return
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("Error getting document ID=%s: %s", id, resp.Status))
		http.Error(w, fmt.Sprintf("Error getting document ID=%s", id), acMux.statusErrorMap(resp.StatusCode))
		return
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(fmt.Sprintf("Error reading response body: %s", err))
		http.Error(w, fmt.Sprintf("Error getting document ID=%s", id), acMux.statusErrorMap(http.StatusInternalServerError))
		return
	}

	// Return the movie details as JSON
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(body)+fmt.Sprintf("%v", acMux.timeModule()))
}

// AutocompleteHandler handles requests to the URL "/autocomplete/$query"
// -- search movies that match the provided query text
// -- return movie json files on success
func (acMux *autocompleteMux) AutocompleteHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the query from the URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		slog.Error("Incorrect number of  path segments")
		http.Error(w, "Incorrect number of  path segments", http.StatusBadRequest)
		return
	}
	query := parts[2]

	// Prepare the Get request
	requestBody := map[string]interface{}{
		"query": map[string]interface{}{
			"simple_query_string": map[string]interface{}{
				"query":  query,
				"fields": []string{"*"},
			},
		},
	}
	requestBodyJson, err := json.Marshal(requestBody)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting movie by query=%s, %s", query, err))
		http.Error(w, fmt.Sprintf("Error getting movie by query=%s", query), acMux.statusErrorMap(http.StatusInternalServerError))
		return
	}
	url := fmt.Sprintf("%s/%s/_search/?pretty", acMux.dbAddr, acMux.movieIndex)
	req, err := http.NewRequest("GET", url, bytes.NewReader(requestBodyJson))
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting movie by query=%s, %s", query, err))
		http.Error(w, fmt.Sprintf("Error getting movie by query=%s", query), acMux.statusErrorMap(http.StatusInternalServerError))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the GET request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting movie by query=%s, %s", query, err))
		http.Error(w, fmt.Sprintf("Error getting movie by query=%s", query), acMux.statusErrorMap(http.StatusInternalServerError))
		return
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("Error getting document query=%s: %s", query, resp.Status))
		http.Error(w, fmt.Sprintf("Error getting document query=%s", query), acMux.statusErrorMap(http.StatusInternalServerError))
		return
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(fmt.Sprintf("Error reading response body: %s", err))
		http.Error(w, fmt.Sprintf("Error getting document query=%s", query), acMux.statusErrorMap(http.StatusInternalServerError))
		return
	}

	// Return the movie details as JSON
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(body))
}

func start() {
	// Create a new mux
	acMux := &autocompleteMux{
		ServeMux:   http.NewServeMux(),
		dbAddr:     fmt.Sprintf("http://%s:8082/db", os.Getenv("LOAD_BALANCER_IP")),
		movieIndex: "movies",
		timeModule: time.Now,
	}

	// Register handlers with the mux
	acMux.HandleFunc("/movies/id/", acMux.GetMovieByIDHandler)
	acMux.HandleFunc("/autocomplete/", acMux.AutocompleteHandler)

	// Start the server with the mux
	log.Fatal(http.ListenAndServe(":8080", acMux))
}

func main() {
	start()
}
