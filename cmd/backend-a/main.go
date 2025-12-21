package main

import (
	"log"
	"net/http"
)

type helloHandler struct{}

func (h *helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from backend A"))
}

func main() {
	mux := http.NewServeMux()

	s := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	mux.Handle("/hello", new(helloHandler))

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
