package main

import (
	"log"
	"net/http"
	"io"
)

func main() {
	mux := http.NewServeMux()

	s := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Add("X-Backend-Name", "backend-B")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from backend B"))
	})

	mux.HandleFunc("/query-params", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Backend-Name", "backend-B")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("list of query params " + r.URL.Query().Encode()))
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Backend-Name", "backend-B")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	})

	mux.HandleFunc("/read-body", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Backend-Name", "backend-B")
		w.WriteHeader(http.StatusOK)

		b, _ := io.ReadAll(r.Body)
		w.Write([]byte(b))
	})

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
