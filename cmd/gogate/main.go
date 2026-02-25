package main

import (
	"log"
	"net/http"
	"strings"

	middleware "github.com/kusiewicz/reverse-proxy/internal/middleware"
	httpproxy "github.com/kusiewicz/reverse-proxy/internal/proxy/http"
)

type gatewayHandler struct {
	sem chan struct{}
}

var cfg = httpproxy.RequestConfig{
	TimeoutInSeconds: 15,
	ConcurrencyLimit: 1,
	MaxRetries:       3,
}

func (g *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/api/a" || strings.HasPrefix(path, "/api/a/") {
		routePrefix := "/api/a"
		if strings.HasPrefix(path, "/api/a/") {
			routePrefix = "/api/a/"
		}

		// wspoldziela semafor - do poprawy
		httpproxy.HandleRequest(w, r, "http://localhost:8081", routePrefix, cfg, g.sem)
		return
	}

	if path == "/api/b" || strings.HasPrefix(path, "/api/b/") {
		routePrefix := "/api/b"
		if strings.HasPrefix(path, "/api/b/") {
			routePrefix = "/api/b/"
		}

		httpproxy.HandleRequest(w, r, "http://localhost:8082", routePrefix, cfg, g.sem)
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

	concurrentRequestsSemaphore := make(chan struct{}, cfg.ConcurrencyLimit)
	for i := 0; i < cfg.ConcurrencyLimit; i++ {
		concurrentRequestsSemaphore <- struct{}{}
	}

	// circuitBreaker := circuitBreaker{
	// 	state:           "Closed",
	// 	errorCounter:    10,
	// 	openTimeSeconds: 30 * time.Second,
	// }

	var h http.Handler
	h = &gatewayHandler{
		sem: concurrentRequestsSemaphore,
	}
	h = middleware.AccessLog(h)
	h = middleware.RequestID(h)
	h = middleware.PanicRecovery(h)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.Handle("/", h)

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
