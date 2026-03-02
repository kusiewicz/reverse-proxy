package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	middleware "github.com/kusiewicz/reverse-proxy/internal/middleware"
	httpproxy "github.com/kusiewicz/reverse-proxy/internal/proxy/http"
)

type gatewayHandler struct {
	configBackendA httpproxy.RequestConfig
	configBackendB httpproxy.RequestConfig
}

func prepareClientConfig(concurrencyLimit int, circuitErrorCounter int, circuitOpenTimeInSeconds int, requestTimeoutInSeconds int, requestMaxRetries int) httpproxy.RequestConfig {
	concurrentRequestSemaphore := make(chan struct{}, concurrencyLimit)

	for i := 0; i < concurrencyLimit; i++ {
		concurrentRequestSemaphore <- struct{}{}
	}

	circuitBreaker := &httpproxy.CircuitBreaker{
		State:              httpproxy.StateClosed,
		ErrorCapWhenOpened: circuitErrorCounter,
		OpenTimeSeconds:    time.Duration(circuitOpenTimeInSeconds) * time.Second,
	}

	return httpproxy.RequestConfig{
		TimeoutInSeconds: requestTimeoutInSeconds,
		MaxRetries:       requestMaxRetries,
		Sem:              concurrentRequestSemaphore,
		CircuitBreaker:   circuitBreaker,
	}
}

func (g *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/api/a" || strings.HasPrefix(path, "/api/a/") {
		routePrefix := "/api/a"
		if strings.HasPrefix(path, "/api/a/") {
			routePrefix = "/api/a/"
		}

		httpproxy.HandleRequest(w, r, "http://localhost:8081", routePrefix, g.configBackendA)
		return
	}

	if path == "/api/b" || strings.HasPrefix(path, "/api/b/") {
		routePrefix := "/api/b"
		if strings.HasPrefix(path, "/api/b/") {
			routePrefix = "/api/b/"
		}

		httpproxy.HandleRequest(w, r, "http://localhost:8082", routePrefix, g.configBackendB)
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

	var configBackendA = prepareClientConfig(5, 2, 10, 10, 5)
	var configBackendB = prepareClientConfig(5, 2, 10, 10, 5)

	var h http.Handler
	h = &gatewayHandler{
		configBackendA: configBackendA,
		configBackendB: configBackendB,
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
