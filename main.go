package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

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

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register handlers with the mux
	mux.HandleFunc("/", DefaultRootHandler)
	mux.HandleFunc("/g", CallExtServiceHandler)

	// Start the server with the mux
	log.Fatal(http.ListenAndServe(":8080", mux))
}
