package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type gatewayHandler struct{}

func cutPrefixPath(path string, prefix string) string {
	new, _ := strings.CutPrefix(path, prefix)

	return new
}

func genereateRequestURL(serverUrl string, path string, query string) string {
	requestUrl := serverUrl + "/" + path

	if len(query) != 0 {
		requestUrl += "?" + query
	}

	return requestUrl
}

func handleRequest(w http.ResponseWriter, r *http.Request, serverURL string, routePrefix string) {
	client := &http.Client{}

	path := r.URL.Path
	query := r.URL.Query().Encode()
	method := r.Method

	strippedPath := cutPrefixPath(path, routePrefix)

	requestURL := genereateRequestURL(serverURL, strippedPath, query)

	req, err := http.NewRequest(method, requestURL, nil)

	fmt.Println("serverURL", serverURL)
	fmt.Println("routePrefix", routePrefix)
	fmt.Println("path", path)
	fmt.Println("requestURL", requestURL)

	if err != nil {
		log.Println("Error when creating request")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println("Error when accessing " + routePrefix)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	headers := resp.Header
	statusCode := resp.StatusCode

	fmt.Println("err", err)

	defer resp.Body.Close()

	for key, values := range headers {
		for _, header := range values {
			w.Header().Add(key, header)
		}
	}

	body, err := io.ReadAll(resp.Body)

	w.WriteHeader(statusCode)
	w.Write(body)
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
