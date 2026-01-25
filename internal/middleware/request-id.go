package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type requestIDHeaderKey struct{}

var requestIDHeaderKeyString = "X-Request-ID"

func RequestIdFrom(ctx context.Context) (string, bool) {
	v := ctx.Value(requestIDHeaderKey{})
	id, ok := v.(string)
	return id, ok
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestIdHeaderValue := r.Header.Get(requestIDHeaderKeyString)

		if requestIdHeaderValue != "" {
			w.Header().Set(requestIDHeaderKeyString, requestIdHeaderValue)
			next.ServeHTTP(w, r)
			return
		}

		requestIdHeaderValue = uuid.New().String()

		ctx := context.WithValue(r.Context(), requestIDHeaderKey{}, requestIdHeaderValue)

		reqWithContext := r.WithContext(ctx)

		r.Header.Set(requestIDHeaderKeyString, requestIdHeaderValue)
		w.Header().Set(requestIDHeaderKeyString, requestIdHeaderValue)

		next.ServeHTTP(w, reqWithContext)
	})
}
