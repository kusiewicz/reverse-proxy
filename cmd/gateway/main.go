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

func (g *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	path := r.URL.Path
	query := r.URL.Query().Encode()
	method := r.Method

	log.Println(path)

	// nameFromQuery := query.Get("name")

	if path == "/api/a" || strings.HasPrefix(path, "/api/a/") {
		// resp, err := http.Get("http://localhost:8081/hello")

		prefixToCut := "/api/a"

		if strings.HasPrefix(path, "/api/a/") {
			prefixToCut = "/api/a/"
		}

		requestRealPath := cutPrefixPath(path, prefixToCut)

		fullUrl := "http://localhost:8081/hello" + requestRealPath + query

		fmt.Println(fullUrl)

		req, err := http.NewRequest(method, fullUrl, nil)

		resp, err := client.Do(req)

		headers := resp.Header
		statusCode := resp.StatusCode

		if err != nil {
			log.Println("Error when accessing /a")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}

		defer resp.Body.Close()

		for key, valuesArray := range headers {
			for _, header := range valuesArray {
				w.Header().Add(key, header)
			}
		}

		body, err := io.ReadAll(resp.Body)

		w.WriteHeader(statusCode)
		w.Write(body)
		return
	}

	if path == "/api/b" || strings.HasPrefix(path, "/api/b/") {
		if method == "GET" {
			resp, err := http.Get("http://localhost:8082/hello")

			headers := resp.Header
			statusCode := resp.StatusCode

			if err != nil {
				log.Println("Error when accessing /a")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}

			defer resp.Body.Close()

			for key, valuesArray := range headers {
				for _, header := range valuesArray {
					w.Header().Add(key, header)
				}
			}

			body, err := io.ReadAll(resp.Body)

			w.WriteHeader(statusCode)
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
