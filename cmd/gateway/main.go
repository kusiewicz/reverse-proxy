package main

import (
	"log"
	"net/http"
	"strings"
)

type gatewayHandler struct{}

func (g *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/api/a" || strings.HasPrefix(path, "/api/a/") {
		log.Println("would forward to backend A")
		w.WriteHeader(http.StatusOK)
		w.Write(([]byte("would forward to backend A")))
		return
	}

	if path == "/api/b" || strings.HasPrefix(path, "/api/b/") {
		log.Println("would forward to backend B")
		w.WriteHeader(http.StatusOK)
		w.Write(([]byte("would forward to backend B")))
		return
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write(([]byte("not found")))
}

func main() {
	mux := http.NewServeMux()

	s := &http.Server{
		Addr:    ":8090",
		Handler: mux,
	}

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.Handle("/", new(gatewayHandler))

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
