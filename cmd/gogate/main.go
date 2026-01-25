package main

import (
	"log"
	"net/http"
	"strings"

	middleware "github.com/kusiewicz/reverse-proxy/internal/middleware"
	httpproxy "github.com/kusiewicz/reverse-proxy/internal/proxy/http"
)

type gatewayHandler struct{}

func (g *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/api/a" || strings.HasPrefix(path, "/api/a/") {
		routePrefix := "/api/a"
		if strings.HasPrefix(path, "/api/a/") {
			routePrefix = "/api/a/"
		}

		httpproxy.HandleRequest(w, r, "http://localhost:8081", routePrefix)
		return
	}

	if path == "/api/b" || strings.HasPrefix(path, "/api/b/") {
		routePrefix := "/api/b"
		if strings.HasPrefix(path, "/api/b/") {
			routePrefix = "/api/b/"
		}

		httpproxy.HandleRequest(w, r, "http://localhost:8082", routePrefix)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write(([]byte("Not found")))
}

func main() {
	mux := http.NewServeMux()
	s := &http.Server{
		Addr:    ":8090",
		Handler: mux,
	}
	handler := new(gatewayHandler)
	handlerWithMiddleware := middleware.RequestID(handler)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.Handle("/", handlerWithMiddleware)

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
