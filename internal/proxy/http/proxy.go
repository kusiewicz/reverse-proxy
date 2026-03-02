package httpproxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type circuitBreakerState string

const (
	StateClosed   circuitBreakerState = "Closed"
	StateHalfOpen circuitBreakerState = "HalfOpen"
	StateOpen     circuitBreakerState = "Open"
)

type CircuitBreaker struct {
	State              circuitBreakerState
	ErrorCapWhenOpened int
	CurrentErrors      int
	OpenTimeSeconds    time.Duration
	Mutex              sync.Mutex
	openAt             time.Time
}

type RequestConfig struct {
	TimeoutInSeconds int
	MaxRetries       int
	Sem              chan struct{}
	CircuitBreaker   *CircuitBreaker
}

var hopByHopHeaders = map[string]struct{}{
	"connection":          {},
	"keep-alive":          {},
	"proxy-authenticate":  {},
	"proxy-authorization": {},
	"te":                  {},
	"trailer":             {},
	"transfer-encoding":   {},
	"upgrade":             {},
}

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

func connectionHeaderTokens(connectionHeaderValue string) map[string]struct{} {
	connectionTokensToDelete := make(map[string]struct{})

	listOfConnectionHeaderValues := strings.Split(connectionHeaderValue, ",")
	for _, v := range listOfConnectionHeaderValues {
		connectionTokensToDelete[strings.ToLower(strings.TrimSpace(v))] = struct{}{}
	}

	return connectionTokensToDelete
}

func logError(proxyError error, routePrefix, serverURL, path, query, method, stage string) {
	log.Printf("Error: %v on stage: %s: route: %s, server: %s, path: %s, query: %s, method: %s", proxyError, stage, routePrefix, serverURL, path, query, method)
}

func getIsRetryable(method string, statusCode int) bool {
	if (statusCode == 502 || statusCode == 503 || statusCode == 504) && (method == "GET" || method == "HEAD") {
		return true
	}
	return false
}

func getShouldBreakRequest(circuitBreakerCfg *CircuitBreaker) bool {
	if circuitBreakerCfg.State == StateOpen {
		if time.Since(circuitBreakerCfg.openAt) > circuitBreakerCfg.OpenTimeSeconds {
			circuitBreakerCfg.State = StateHalfOpen
			return false
		}
		if time.Since(circuitBreakerCfg.openAt) < circuitBreakerCfg.OpenTimeSeconds {
			return true
		}
	} else {
		if circuitBreakerCfg.CurrentErrors >= circuitBreakerCfg.ErrorCapWhenOpened {
			circuitBreakerCfg.State = StateOpen
			circuitBreakerCfg.openAt = time.Now()
			return true
		}
	}
	return false
}

func HandleRequest(w http.ResponseWriter, r *http.Request, serverURL string, routePrefix string, cfg RequestConfig) {
	timeout := time.Duration(cfg.TimeoutInSeconds) * time.Second
	client := &http.Client{}

	path := r.URL.Path
	query := r.URL.Query().Encode()
	method := r.Method

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
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

	requestConnectionHeadersToDelete := make(map[string]struct{})

	connectionHeader := r.Header.Get("Connection")

	if connectionHeader != "" {
		requestConnectionHeadersToDelete = connectionHeaderTokens(connectionHeader)
	}

	for key, values := range requestHeaders {
		loweredKey := strings.ToLower(key)

		if _, ok := requestConnectionHeadersToDelete[loweredKey]; ok {
			continue
		}
		if _, ok := hopByHopHeaders[loweredKey]; ok {
			continue
		}

		for _, header := range values {
			req.Header.Add(key, header)
		}
	}

	shouldBreakRequest := getShouldBreakRequest(cfg.CircuitBreaker)

	if shouldBreakRequest {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Circuit Breaker err"))
		return
	}

	<-cfg.Sem
	defer func() { cfg.Sem <- struct{}{} }()

	// select {
	// case <-requestSemaphore:
	// 	defer func() { requestSemaphore <- struct{}{} }()
	// default:
	// 	w.WriteHeader(http.StatusGatewayTimeout)
	// 	return
	// }

	resp, err := client.Do(req)

	if err == nil && getIsRetryable(req.Method, resp.StatusCode) && cfg.MaxRetries > 0 {
		cfg.CircuitBreaker.CurrentErrors++
		for i := 0; i < cfg.MaxRetries; i++ {
			select {
			case <-time.After(500 * time.Millisecond):
				if resp != nil {
					resp.Body.Close()
				}
				shouldBreakRequest := getShouldBreakRequest(cfg.CircuitBreaker)
				if shouldBreakRequest {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte("Circuit Breaker err"))
					return
				}
				resp, err = client.Do(req)
			case <-ctx.Done():
				if resp != nil {
					resp.Body.Close()
				}
				return
			}

			if err == nil && !getIsRetryable(req.Method, resp.StatusCode) {
				cfg.CircuitBreaker.State = StateClosed
				cfg.CircuitBreaker.openAt = time.Time{}
				cfg.CircuitBreaker.CurrentErrors = 0
				break
			}

			if err != nil || getIsRetryable(req.Method, resp.StatusCode) {
				fmt.Println("Increment circuit")
				cfg.CircuitBreaker.CurrentErrors += 1
			}
		}
	}

	if err != nil {
		cfg.CircuitBreaker.CurrentErrors += 1
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
	cfg.CircuitBreaker.CurrentErrors = 0
	cfg.CircuitBreaker.State = StateClosed
	cfg.CircuitBreaker.openAt = time.Time{}

	defer resp.Body.Close()

	responseConnectionHeadersToDelete := make(map[string]struct{})

	connectionHeader = responseHeaders.Get("Connection")

	if connectionHeader != "" {
		responseConnectionHeadersToDelete = connectionHeaderTokens(connectionHeader)
	}

	for key, values := range responseHeaders {
		loweredKey := strings.ToLower(key)

		if _, ok := responseConnectionHeadersToDelete[loweredKey]; ok {
			continue
		}
		if _, ok := hopByHopHeaders[loweredKey]; ok {
			continue
		}

		for _, header := range values {
			w.Header().Add(key, header)
		}
	}

	w.WriteHeader(responseStatusCode)

	// io.Copy streaming instead of ReadAll
	if _, err := io.Copy(w, resp.Body); err != nil {
		logError(err, routePrefix, serverURL, path, query, method, "response streaming")
		return
	}
}
