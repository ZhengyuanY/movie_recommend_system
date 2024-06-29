package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func DefaultRootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, server UP!")
}

func CallExtServiceHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://www.google.com/ncr")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close() // Ensure the response body is closed

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Fprintln(w, string(body))
}

func main() {
	http.HandleFunc("/", DefaultRootHandler)
	http.HandleFunc("/g", CallExtServiceHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
