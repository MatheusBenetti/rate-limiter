package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiterMapIP     map[string]*rate.Limiter
	limiterMapToken  map[string]*rate.Limiter
	redisClient      *redis.Client
	blockDuration    time.Duration
	rateLimitPerPage int
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiterMapIP:     make(map[string]*rate.Limiter),
		limiterMapToken:  make(map[string]*rate.Limiter),
		redisClient:      getRedisClient(),
		blockDuration:    5 * time.Minute, // Tempo de bloqueio em caso de exceder as requisições
		rateLimitPerPage: 5,               // Número máximo de requisições permitidas por segundo
	}
}

func (rl *RateLimiter) getLimiterIP(ip string) *rate.Limiter {
	rl.redisClient.Set(context.Background(), "ip:"+ip, 0, rl.blockDuration)
	limiter, exists := rl.limiterMapIP[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.rateLimitPerPage), rl.rateLimitPerPage)
		rl.limiterMapIP[ip] = limiter
	}
	return limiter
}

func (rl *RateLimiter) getLimiterToken(token string) *rate.Limiter {
	rl.redisClient.Set(context.Background(), "token:"+token, 0, rl.blockDuration)
	limiter, exists := rl.limiterMapToken[token]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.rateLimitPerPage), rl.rateLimitPerPage)
		rl.limiterMapToken[token] = limiter
	}
	return limiter
}

func (rl *RateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		token := r.Header.Get("API_KEY")

		limiterIP := rl.getLimiterIP(ip)
		limiterToken := rl.getLimiterToken(token)

		if !limiterIP.Allow() && !limiterToken.Allow() {
			http.Error(w, "You have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	rl := NewRateLimiter()

	router := mux.NewRouter()
	router.Use(rl.middleware)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // use default DB
	})
	return rdb
}
