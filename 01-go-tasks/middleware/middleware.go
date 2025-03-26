package middleware

import (
	"log"
	"net/http"
	"time"
)

// Middleware defines a function to process HTTP requests
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Chain applies middlewares to a http.HandlerFunc
func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

// Logging logs request details
func Logging() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log.Printf("Started %s %s", r.Method, r.URL.Path)

			f(w, r)

			log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
		}
	}
}

// ContentTypeJSON sets the Content-Type header to application/json
func ContentTypeJSON() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			f(w, r)
		}
	}
}
