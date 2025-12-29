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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from backend B"))
	})

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
