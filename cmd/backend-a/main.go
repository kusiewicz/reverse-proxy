package main

import (
	"log"
	"net/http"
)

type helloHandler struct{}
type errorHandler struct{}

func (h *helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Backend-Name", "backend-a")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from backend A"))
}

func (h *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Backend-Name", "backend-a")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func main() {
	mux := http.NewServeMux()

	s := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	mux.Handle("/hello", new(helloHandler))
	mux.Handle("/error", new(errorHandler))

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
