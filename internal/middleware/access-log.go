package middleware

import (
	"log"
	"net/http"
	"time"
)

type ResponseWriterWrapper struct {
	inner  http.ResponseWriter
	status int
	bytes  int
}

func (rw *ResponseWriterWrapper) Header() http.Header {
	return rw.inner.Header()
}

func (rw *ResponseWriterWrapper) Write(p []byte) (int, error) {
	if rw.status == 0 {
		rw.status = 200
	}
	n, err := rw.inner.Write(p)
	rw.bytes += n
	return n, err
}

func (rw *ResponseWriterWrapper) WriteHeader(statusCode int) {
	if rw.status == 0 {
		rw.status = statusCode
	}
	rw.inner.WriteHeader(statusCode)
}

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		method := r.Method
		path := r.URL.Path
		requestId, ok := RequestIdFrom(r.Context())

		if ok != true {
			requestId = r.Header.Get(RequestIDHeaderName)
		}

		rw := &ResponseWriterWrapper{inner: w}

		next.ServeHTTP(rw, r)

		latency := time.Since(start)
		status := rw.status
		if status == 0 {
			status = 200
		}
		bytes := rw.bytes

		log.Printf("method=%s path=%s requestId=%s latency=%s status=%d bytes=%d", method, path, requestId, latency, status, bytes)
	})
}
