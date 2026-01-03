package main

import (
	"fmt"
	"log"
	"net/http"
)

type helloHandler struct{}
type errorHandler struct{}
type queryParamsHandler struct{}

func (h *helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Add("X-Backend-Name", "backend-a")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from backend A"))
}

func (h *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Backend-Name", "backend-a")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func (q *queryParamsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Backend-Name", "backend-a")
	w.WriteHeader(http.StatusOK)

	fmt.Println(r.URL.Query().Encode())

	w.Write([]byte("list of query params " + r.URL.Query().Encode()))
}

func main() {
	mux := http.NewServeMux()

	s := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	mux.Handle("/", new(helloHandler))
	mux.Handle("/error", new(errorHandler))
	mux.Handle("/query-params", new(queryParamsHandler))

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
