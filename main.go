package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiterMapIP    map[string]*rate.Limiter
	limiterMapToken map[string]*rate.Limiter
	redisClient     *redis.Client
	blockDuration   time.Duration
}

func NewRateLimiter() *RateLimiter {
	rdb := getRedisClient()
	return &RateLimiter{
		limiterMapIP:    make(map[string]*rate.Limiter),
		limiterMapToken: make(map[string]*rate.Limiter),
		redisClient:     rdb,
		blockDuration:   5 * time.Minute, // Tempo de bloqueio em caso de exceder as requisições
	}
}

func (rl *RateLimiter) getLimiterIP(ip string) *rate.Limiter {
	rl.redisClient.Set(context.Background(), "ip:"+ip, 0, rl.blockDuration)
	limiter, exists := rl.limiterMapIP[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(getRateLimit("IP")), 10)
		rl.limiterMapIP[ip] = limiter
	}
	return limiter
}

func (rl *RateLimiter) getLimiterToken(token string) *rate.Limiter {
	rl.redisClient.Set(context.Background(), "token:"+token, 0, rl.blockDuration)
	limiter, exists := rl.limiterMapToken[token]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(getRateLimit("Token")), 20)
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

func getRateLimit(category string) float64 {
	if category == "IP" {
		os.Setenv(category+"_RATE_LIMIT", "10")
	} else if category == "Token" {
		os.Setenv(category+"_RATE_LIMIT", "20")
	}

	limit, err := strconv.ParseFloat(os.Getenv(category+"_RATE_LIMIT"), 64)
	if err != nil {
		panic(fmt.Sprintf("Error parsing rate limit for %s: %v", category, err))
	}
	return limit
}

func getRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // use default DB
	})
	return rdb
}

func main() {
	rl := NewRateLimiter()

	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	http.Handle("/", rl.middleware(router))
	log.Println("Server started on port 8080.")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
