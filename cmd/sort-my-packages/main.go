package main

import (
	"log"
	"net/http"

	"github.com/APoniatowski/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	mux.HandleFunc("POST /calculate-packs", handlers.CalculatePacks)
	mux.HandleFunc("POST /set-pack-sizes", handlers.SetPackSizes)

	log.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
