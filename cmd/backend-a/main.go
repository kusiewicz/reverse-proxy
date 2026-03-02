package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

type helloHandler struct{}
type errorHandler struct{}
type queryParamsHandler struct{}
type checkBodyHandler struct{}

func (h *helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	zerotoone := rand.Float64()

	fmt.Println(zerotoone)

	if zerotoone < 0.8 {
		w.Header().Add("X-Backend-Name", "backend-a")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from backend A"))
	} else {
		w.Header().Add("X-Backend-Name", "backend-a")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Error backend A"))
	}
}

func (h *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Backend-Name", "backend-a")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func (q *queryParamsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("received header", r.Header)
	sleep, _ := strconv.Atoi(r.URL.Query().Get("sleep"))

	time.Sleep(time.Duration(sleep) * time.Second)

	w.Header().Add("X-Backend-Name", "backend-a")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("list of query params " + r.URL.Query().Encode()))
}

func (c *checkBodyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Backend-Name", "backend-a")
	w.WriteHeader(http.StatusOK)
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
	mux.Handle("/check-body", new(checkBodyHandler))

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
