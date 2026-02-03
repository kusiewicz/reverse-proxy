package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			v := recover()

			if v != nil {
				log.Printf("panic: %v", v)
				log.Printf("stack:\n%s", debug.Stack())

				w.WriteHeader(500)
				w.Write([]byte("Internal error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
