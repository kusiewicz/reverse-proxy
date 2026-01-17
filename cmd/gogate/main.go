package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type gatewayHandler struct{}

func cutPrefixPath(path string, prefix string) string {
	new, _ := strings.CutPrefix(path, prefix)

	return new
}

func generateRequestURL(serverURL string, path string, query string) string {
	requestURL := serverURL + "/" + path

	if len(query) != 0 {
		requestURL += "?" + query
	}

	return requestURL
}

func logError(proxyError error, routePrefix, serverURL, path, query, method, stage string) {
	log.Printf("Error: %v on stage: %s: route: %s, server: %s, path: %s, query: %s, method: %s", proxyError, stage, routePrefix, serverURL, path, query, method)
}

func handleRequest(w http.ResponseWriter, r *http.Request, serverURL string, routePrefix string) {
	defaultTimeout := 15 * time.Second
	client := &http.Client{}

	path := r.URL.Path
	query := r.URL.Query().Encode()
	method := r.Method

	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeout)
	defer cancel()

	strippedPath := cutPrefixPath(path, routePrefix)

	requestURL := generateRequestURL(serverURL, strippedPath, query)
	req, err := http.NewRequestWithContext(ctx, method, requestURL, r.Body)

	requestHeaders := r.Header

	if err != nil {
		logError(err, routePrefix, serverURL, path, query, method, "request creating")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	for key, values := range requestHeaders {
		for _, header := range values {
			req.Header.Add(key, header)
		}
	}

	resp, err := client.Do(req)

	if err != nil {
		errorStatusCode := http.StatusBadGateway
		errorMessage := "Bad Gateway"

		if ctx.Err() == context.DeadlineExceeded {
			errorStatusCode = http.StatusGatewayTimeout
			errorMessage = "Gateway Timeout"
		}

		if ctx.Err() == context.Canceled {
			logError(err, routePrefix, serverURL, path, query, method, "request")
			return
		}

		logError(err, routePrefix, serverURL, path, query, method, "request")
		w.WriteHeader(errorStatusCode)
		w.Write([]byte(errorMessage))
		return
	}

	responseHeaders := resp.Header
	responseStatusCode := resp.StatusCode

	defer resp.Body.Close()

	for key, values := range responseHeaders {
		for _, header := range values {
			w.Header().Add(key, header)
		}
	}

	w.WriteHeader(responseStatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		logError(err, routePrefix, serverURL, path, query, method, "response streaming")
		return
	}
}

func (g *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	if path == "/api/a" || strings.HasPrefix(path, "/api/a/") {
		routePrefix := "/api/a"
		if strings.HasPrefix(path, "/api/a/") {
			routePrefix = "/api/a/"
		}

		handleRequest(w, r, "http://localhost:8081", routePrefix)
		return
	}

	if path == "/api/b" || strings.HasPrefix(path, "/api/b/") {
		routePrefix := "/api/b"
		if strings.HasPrefix(path, "/api/b/") {
			routePrefix = "/api/b/"
		}

		handleRequest(w, r, "http://localhost:8082", routePrefix)
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

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.Handle("/", new(gatewayHandler))

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
