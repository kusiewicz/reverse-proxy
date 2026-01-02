package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	s := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Backend-Name", "backend-B")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from backend B"))
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Backend-Name", "backend-B")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	})

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
