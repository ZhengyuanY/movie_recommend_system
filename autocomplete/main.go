package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var dbAddr string = fmt.Sprintf("http://%s:8082/db", os.Getenv("LOAD_BALANCER_IP"))

// DefaultRootHandler handles requests to the root URL "/"
func DefaultRootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, server UP!")
}

// CallExtServiceHandler handles requests to the URL "/g"
func CallExtServiceHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://www.google.com/ncr")
	if err != nil {
		http.Error(w, "Error calling external service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close() // Ensure the response body is closed

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response body", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, string(body))
}

// GetMovieByID queries a movie from elasticsearch by an id
func GetMovieByID(index, id string) (string, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/%s/_doc/%s?pretty", dbAddr, index, id)

	// Send the GET request
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error getting document ID=%s: %s", id, resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// MovieHandler handles requests to the URL "/movies/$id"
func MovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Movie ID not provided", http.StatusBadRequest)
		return
	}
	id := parts[2]

	// Call GetMovieByID to get the movie details
	movieJSON, err := GetMovieByID("movies", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting movie: %s", err), http.StatusInternalServerError)
		return
	}

	// Return the movie details as JSON
	fmt.Fprintln(w, movieJSON)
}

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register handlers with the mux
	mux.HandleFunc("/", DefaultRootHandler)
	mux.HandleFunc("/g", CallExtServiceHandler)
	mux.HandleFunc("/movies/", MovieHandler)

	// Start the server with the mux
	log.Fatal(http.ListenAndServe(":8080", mux))
}
