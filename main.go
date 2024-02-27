package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

func main() {
	r := chi.NewRouter()
	r.Use(SimpleAuthMiddleware)

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, 1*time.Minute))

		r.Use(httprate.Limit(20, 1*time.Minute, httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		})))

	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("."))
	})

	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", r)
}

func SimpleAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("API_KEY")
		if apiKey != "secret" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
