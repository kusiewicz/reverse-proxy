package main

import (
	"io"
	"log"
	"net/http"
	"strings"
)

type gatewayHandler struct{}

func (g *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// query := r.URL.Query()
	method := r.Method

	// nameFromQuery := query.Get("name")

	if path == "/api/a" || strings.HasPrefix(path, "/api/a/") {
		if method == "GET" {
			resp, err := http.Get("http://localhost:8081/hello")

			if err != nil {
				log.Println("Error when accessing /a")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}

			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)

			w.WriteHeader(http.StatusOK)
			w.Write(body)
			return
		}
	}

	if path == "/api/b" || strings.HasPrefix(path, "/api/b/") {
		if method == "GET" {
			resp, err := http.Get("http://localhost:8082/hello")

			if err != nil {
				log.Println("Error when accessing /a")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}

			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)

			w.WriteHeader(http.StatusOK)
			w.Write(body)
			return
		}
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
