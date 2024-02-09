package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"gopkg.in/redis.v5"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(rateLimiter(10, time.Minute))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the server!\n"))
	})

	r.Get("/limited-ip", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Limited by IP\n"))
	})

	r.Get("/limited-token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("API_KEY", "my-token")
		w.Write([]byte("Limited by Token\n"))
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}

func rateLimiter(maxRequests int, duration time.Duration) func(next http.Handler) http.Handler {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := httprate.LimitByRealIP(r)
			token := r.Header.Get("API_KEY")

			key := fmt.Sprintf("rate-limiter:%s", ip)
			if token != "" {
				key = fmt.Sprintf("rate-limiter:%s:%s", token, ip)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, "key", key)

			limiter := time.NewTicker(duration)
			defer limiter.Stop()

			var remaining int
			var lastAccess time.Time

			for {
				select {
				case <-limiter.C:
					remaining, _ = getRemaining(client, key)
					if remaining >= maxRequests {
						http.Error(w, "Too many requests", http.StatusTooManyRequests)
						return
					}

					lastAccess = time.Now()
					setRemaining(client, key, maxRequests-remaining)
					break
				case <-ctx.Done():
					return
				}

				if remaining == 0 || time.Since(lastAccess) > duration {
					break
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getRemaining(client *redis.Client, key string) (int, error) {
	val, err := client.Get(key).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	}

	if val == "" {
		return maxRequests, nil
	}

	remaining, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return remaining, nil
}

func setRemaining(client *redis.Client, key string, remaining int) error {
	return client.Set(key, remaining, 0).Err()
}

const maxRequests = 10
