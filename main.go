package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

type RateLimiter struct {
	redisClient *redis.Client
	ipLimits    map[string]int
	tokenLimits map[string]int
	mu          sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &RateLimiter{
		redisClient: redisClient,
		ipLimits:    make(map[string]int),
		tokenLimits: make(map[string]int),
	}
}

func (rl *RateLimiter) LimitIP(ip string) bool {
	ctx := context.Background()
	key := "ip:" + ip
	limit, err := rl.redisClient.Get(ctx, key).Result()
	if err != nil {
		return false
	}

	maxRequests, _ := strconv.Atoi(limit)
	now := time.Now().Unix()
	keyTimestamp := key + ":timestamp"
	timestamp, err := rl.redisClient.Get(ctx, keyTimestamp).Int64()
	if err != nil {
		timestamp = now
	}

	elapsedTime := now - timestamp
	if elapsedTime > 1 {
		rl.redisClient.Set(ctx, keyTimestamp, now, time.Minute)
		rl.redisClient.Set(ctx, key, 1, time.Minute)
		return true
	}

	count, err := rl.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return false
	}

	if count > int64(maxRequests) {
		return false
	}

	return true
}

func (rl *RateLimiter) LimitToken(token string) bool {
	ctx := context.Background()
	key := "token:" + token
	limit, err := rl.redisClient.Get(ctx, key).Result()
	if err != nil {
		return false
	}

	maxRequests, _ := strconv.Atoi(limit)
	now := time.Now().Unix()
	keyTimestamp := key + ":timestamp"
	timestamp, err := rl.redisClient.Get(ctx, keyTimestamp).Int64()
	if err != nil {
		timestamp = now
	}

	elapsedTime := now - timestamp
	if elapsedTime > 1 {
		rl.redisClient.Set(ctx, keyTimestamp, now, time.Minute)
		rl.redisClient.Set(ctx, key, 1, time.Minute)
		return true
	}

	count, err := rl.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return false
	}

	if count > int64(maxRequests) {
		return false
	}

	return true
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		token := r.Header.Get("API_KEY")

		if rl.LimitIP(ip) && rl.LimitToken(token) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "You have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		}
	})
}

func main() {
	rateLimiter := NewRateLimiter()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	http.Handle("/", rateLimiter.Middleware(http.DefaultServeMux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server started at :" + port)
	http.ListenAndServe(":"+port, nil)
}
